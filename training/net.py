"""The network: a small AlphaZero-style ResNet, ego-centric in, one seat out.

Input  : PLANES x board x board, one seat's view (see encode.py)
Policy : 4 logits, that seat's move distribution
Value  : 1 scalar in [-1, 1], that seat's expected placement score

One scalar rather than a 4-vector because the mode is not zero-sum: each seat is
evaluated from its own perspective and the four values need not sum to anything.

Run this file to export an ONNX model for Go to serve:

    python training/net.py --out runs/dev/champion.onnx --board 11
"""

import argparse

import torch
import torch.nn as nn

from encode import PLANES


class Block(nn.Module):
    def __init__(self, filters):
        super().__init__()
        self.c1 = nn.Conv2d(filters, filters, 3, padding=1, bias=False)
        self.b1 = nn.BatchNorm2d(filters)
        self.c2 = nn.Conv2d(filters, filters, 3, padding=1, bias=False)
        self.b2 = nn.BatchNorm2d(filters)

    def forward(self, x):
        y = torch.relu(self.b1(self.c1(x)))
        return torch.relu(x + self.b2(self.c2(y)))


class Net(nn.Module):
    def __init__(self, board=11, filters=64, blocks=4):
        super().__init__()
        self.board = board
        self.stem = nn.Sequential(
            nn.Conv2d(PLANES, filters, 3, padding=1, bias=False),
            nn.BatchNorm2d(filters),
            nn.ReLU(),
        )
        self.trunk = nn.Sequential(*[Block(filters) for _ in range(blocks)])

        # Both heads squeeze to a couple of planes before flattening, which is
        # what keeps the fully connected layers small.
        self.policy = nn.Sequential(
            nn.Conv2d(filters, 2, 1, bias=False), nn.BatchNorm2d(2), nn.ReLU(), nn.Flatten(),
            nn.Linear(2 * board * board, 4),
        )
        self.value = nn.Sequential(
            nn.Conv2d(filters, 1, 1, bias=False), nn.BatchNorm2d(1), nn.ReLU(), nn.Flatten(),
            nn.Linear(board * board, 64), nn.ReLU(),
            nn.Linear(64, 1), nn.Tanh(),
        )

    def forward(self, x):
        x = self.trunk(self.stem(x))
        # Logits, not probabilities: training wants them raw for cross-entropy,
        # and search softmaxes them itself.
        return self.policy(x), self.value(x).squeeze(-1)


def export(net, path, board):
    """Writes an ONNX model with a dynamic batch axis.

    Batching matters: a node is evaluated once per seat, so inference goes out
    as a batch of 4, and self-play batches further across simulations.
    """
    net.eval()
    torch.onnx.export(
        net,
        torch.zeros(1, PLANES, board, board),
        path,
        input_names=["board"],
        output_names=["policy", "value"],
        dynamic_axes={"board": {0: "batch"}, "policy": {0: "batch"}, "value": {0: "batch"}},
        opset_version=17,
    )


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("--out", default="runs/dev/champion.onnx")
    ap.add_argument("--board", type=int, default=11)
    ap.add_argument("--filters", type=int, default=64)
    ap.add_argument("--blocks", type=int, default=4)
    ap.add_argument("--seed", type=int, default=0)
    args = ap.parse_args()

    import os

    torch.manual_seed(args.seed)
    net = Net(board=args.board, filters=args.filters, blocks=args.blocks)
    os.makedirs(os.path.dirname(args.out) or ".", exist_ok=True)
    export(net, args.out, args.board)
    print(f"{sum(p.numel() for p in net.parameters())} parameters -> {args.out}")
