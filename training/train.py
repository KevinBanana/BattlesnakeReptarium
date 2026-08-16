"""One training pass: replay buffer in, candidate network out.

    python training/train.py --run runs/dev --out runs/dev/iter-0007

Training continues from the champion's weights rather than starting fresh each
iteration - the network is meant to improve on itself, not to relearn the game
every time.

    python training/train.py --check     # verify the D4 augmentation
"""

import argparse
import os
import sys

import numpy as np
import torch
import torch.nn.functional as F

from encode import PLANE_ENEMY_HEADS, PLANE_OWN_HEAD, PLANES
from net import Net, export
from samples import read_iterations, to_planes


def augment(planes, policy, k, flip):
    """Applies one of the 8 board symmetries to inputs and policy together.

    With no food and no hazards, a constrictor position is symmetric under all
    8 rotations and reflections with no caveats, so every sample is worth 8.
    Rotating the board without rotating the policy target teaches the network
    that up is sometimes left, which is worse than no augmentation at all.

    The policy transform is a roll because AllMoves is already in rotational
    order: Up, Left, Down, Right turns counterclockwise each step. The roll is
    *negative* because cells are indexed y*w+x with y running up the board,
    which makes np.rot90 on those axes send (y, x) to (n-1-x, y) - a clockwise
    quarter turn, the other way. Getting that backwards teaches the network that
    up is sometimes down, which is worse than no augmentation at all, and is
    exactly what self_check pins.
    """
    if flip:
        planes = np.flip(planes, axis=-1)  # mirror x
        policy = policy[..., [0, 3, 2, 1]]  # Left and Right swap, Up and Down stay
    if k:
        planes = np.rot90(planes, k, axes=(-2, -1))
        policy = np.roll(policy, -k, axis=-1)
    return np.ascontiguousarray(planes), np.ascontiguousarray(policy)


def self_check():
    """Asserts the policy target rotates with the board."""
    board = 5
    planes = np.zeros((PLANES, board, board), dtype=np.float32)
    planes[PLANE_OWN_HEAD, 0, 2] = 1  # bottom edge, middle column
    planes[PLANE_ENEMY_HEADS, 4, 0] = 1  # top-left corner
    policy = np.array([1, 0, 0, 0], dtype=np.float32)  # all weight on Up

    # One rot90 sends the bottom edge to the left edge - (y=0, x=2) lands at
    # (y=2, x=0) - so a move that was Up is now Right.
    rotated, rolled = augment(planes, policy, k=1, flip=False)
    assert rotated[PLANE_OWN_HEAD, 2, 0] == 1, np.argwhere(rotated[PLANE_OWN_HEAD])
    # (y=4, x=0) lands at (y=4, x=4): the map is (y, x) -> (n-1-x, y).
    assert rotated[PLANE_ENEMY_HEADS, 4, 4] == 1, np.argwhere(rotated[PLANE_ENEMY_HEADS])
    assert rolled.tolist() == [0, 0, 0, 1], rolled

    # Four rotations are the identity, for both.
    same, samep = augment(planes, policy, k=4, flip=False)
    assert np.array_equal(same, planes) and np.array_equal(samep, policy)

    # A mirror leaves Up alone and swaps Left with Right.
    _, mirrored = augment(planes, np.array([0, 1, 0, 0], dtype=np.float32), k=0, flip=True)
    assert mirrored.tolist() == [0, 0, 0, 1], mirrored

    print("augmentation self-check passed")


def load_net(run_dir, board, filters, blocks, device):
    """Continues from the champion, or starts fresh on the first iteration."""
    net = Net(board=board, filters=filters, blocks=blocks).to(device)
    champion = os.path.join(run_dir, "champion.pt")
    if os.path.exists(champion):
        net.load_state_dict(torch.load(champion, map_location=device))
        print(f"  continuing from {champion}")
    else:
        print("  no champion yet, starting from random weights")
    return net


def train(args):
    device = "cuda" if torch.cuda.is_available() else "cpu"

    records, (w, h, seats), iterations = read_iterations(args.run, args.keep)
    print(f"  buffer: {len(records)} samples from {len(iterations)} iterations ({iterations[0]}..{iterations[-1]})")

    planes = to_planes(records, w, h)
    policy = np.ascontiguousarray(records["policy"])
    value = np.ascontiguousarray(records["value"])

    net = load_net(args.run, w, args.filters, args.blocks, device)
    optimizer = torch.optim.Adam(net.parameters(), lr=args.lr, weight_decay=1e-4)
    net.train()

    order = np.arange(len(records))
    steps = policy_loss = value_loss = 0.0
    for epoch in range(args.epochs):
        np.random.shuffle(order)
        for start in range(0, len(order), args.batch):
            idx = order[start : start + args.batch]

            # One symmetry per batch rather than eight copies of it: the samples
            # are shuffled anyway, so over an epoch every orientation shows up,
            # and the buffer stays the size it already is.
            k, flip = np.random.randint(4), bool(np.random.randint(2))
            x, p = augment(planes[idx], policy[idx], k, flip)

            x = torch.from_numpy(x).to(device)
            p = torch.from_numpy(p).to(device)
            v = torch.from_numpy(value[idx]).to(device)

            logits, predicted = net(x)
            # Soft targets, so this is cross-entropy against the visit
            # distribution rather than against a single "correct" move.
            loss_p = -(p * F.log_softmax(logits, dim=1)).sum(dim=1).mean()
            loss_v = F.mse_loss(predicted, v)
            loss = loss_p + loss_v

            optimizer.zero_grad(set_to_none=True)
            loss.backward()
            optimizer.step()

            policy_loss += loss_p.item()
            value_loss += loss_v.item()
            steps += 1

        print(f"  epoch {epoch + 1}/{args.epochs}  policy {policy_loss / steps:.4f}  value {value_loss / steps:.4f}")

    os.makedirs(args.out, exist_ok=True)
    weights = os.path.join(args.out, "candidate.pt")
    model = os.path.join(args.out, "candidate.onnx")

    net.eval()
    torch.save(net.state_dict(), weights)
    export(net.cpu(), model, w)
    print(f"  wrote {model}")

    return {"samples": len(records), "policy_loss": policy_loss / steps, "value_loss": value_loss / steps}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--run", help="run directory holding iter-*/games.bin")
    ap.add_argument("--out", help="iteration directory to write candidate.onnx into")
    ap.add_argument("--keep", type=int, default=10, help="iterations of games in the replay buffer")
    ap.add_argument("--epochs", type=int, default=1)
    ap.add_argument("--batch", type=int, default=512)
    ap.add_argument("--lr", type=float, default=1e-3)
    ap.add_argument("--filters", type=int, default=64)
    ap.add_argument("--blocks", type=int, default=4)
    ap.add_argument("--check", action="store_true", help="run the augmentation self-check and exit")
    args = ap.parse_args()

    if args.check:
        self_check()
        return
    if not args.run or not args.out:
        sys.exit("--run and --out are required")

    train(args)


if __name__ == "__main__":
    main()
