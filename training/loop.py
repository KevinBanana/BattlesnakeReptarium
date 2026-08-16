"""The training loop: self-play, train, evaluate, promote, repeat.

    python training/loop.py --run runs/dev --iterations 100

Stop it whenever you like - Ctrl-C, a reboot, a power cut - and run the same
command again. It picks up where it left off.

There is no checkpoint file, because the run directory *is* the checkpoint:

    runs/dev/
      champion.onnx        what self-play currently plays
      champion.pt          the same weights, for training to continue from
      log.jsonl            one line per finished iteration
      iter-0000/
        games.bin          self-play records      <- step 1 done
        game.html          a game to watch
        candidate.onnx     trained network        <- step 2 done
        eval.json          how it did             <- step 3 done
      iter-0001/
        ...

Each step checks whether its output already exists and skips if so, so resuming
costs nothing and a half-finished iteration finishes rather than restarting.
Self-play writes games.bin under a .partial name and renames it, so an
interrupted run never leaves a truncated file that looks complete.
"""

import argparse
import glob
import json
import os
import shutil
import subprocess
import sys
import time

import torch

from net import Net, export

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


def log(message):
    print(f"[{time.strftime('%H:%M:%S')}] {message}", flush=True)


def iteration_dir(run, i):
    return os.path.join(run, f"iter-{i:04d}")


def completed_iterations(run):
    """How many iterations are fully done - the resume point."""
    done = 0
    while os.path.exists(os.path.join(iteration_dir(run, done), "eval.json")):
        done += 1
    return done


def find_onnxruntime(cuda):
    """Points the Go side at the right ONNX Runtime library, and at CUDA's.

    Two builds live in third_party: the CPU one, and the GPU one whose provider
    library needs cuBLAS, cuDNN and the CUDA runtime on PATH. Those come from
    the nvidia-* wheels in the venv - torch's copies are a different version and
    load with "the specified procedure could not be found".

    Set ONNXRUNTIME_LIB yourself to override. Without this the loop would only
    run from a shell that had already exported it, which is a poor property for
    the one script you are meant to be able to just run.
    """
    if cuda:
        nvidia = glob.glob(os.path.join(ROOT, ".venv", "Lib", "site-packages", "nvidia", "*", "bin"))
        nvidia += glob.glob(os.path.join(ROOT, ".venv", "lib", "python*", "site-packages", "nvidia", "*", "lib"))
        if nvidia:
            os.environ["PATH"] = os.pathsep.join(nvidia) + os.pathsep + os.environ["PATH"]

    if os.environ.get("ONNXRUNTIME_LIB"):
        return

    want = "onnxruntime-win-x64-gpu*" if cuda else "onnxruntime-win-x64-1*"
    for name in ("onnxruntime.dll", "libonnxruntime.so", "libonnxruntime.dylib"):
        found = glob.glob(os.path.join(ROOT, "third_party", want, "lib", name))
        if found:
            os.environ["ONNXRUNTIME_LIB"] = found[0]
            return
    sys.exit(f"no ONNX Runtime library matching {want} under third_party/")


def shell(args, label):
    """Runs a subprocess, letting its output through so progress is visible."""
    result = subprocess.run(args, cwd=ROOT)
    if result.returncode != 0:
        sys.exit(f"{label} failed with exit code {result.returncode}")


def ensure_champion(args):
    """Creates the starting network if this is a fresh run."""
    champion = os.path.join(args.run, "champion.onnx")
    if os.path.exists(champion):
        return

    os.makedirs(args.run, exist_ok=True)
    log(f"no champion yet: initialising {args.blocks}x{args.filters} network on a {args.board}x{args.board} board")

    torch.manual_seed(args.seed)
    net = Net(board=args.board, filters=args.filters, blocks=args.blocks).eval()
    torch.save(net.state_dict(), os.path.join(args.run, "champion.pt"))
    export(net, champion, args.board)
    log(f"  {sum(p.numel() for p in net.parameters())} parameters")


def self_play(args, out):
    games = os.path.join(out, "games.bin")
    if os.path.exists(games):
        log("  self-play already done, skipping")
        return

    shell(
        [
            "go", "run", "./cmd/selfplay",
            "-model", os.path.join(args.run, "champion.onnx"),
            "-board", str(args.board),
            "-games", str(args.games),
            "-sims", str(args.sims),
            "-samples", games,
            "-html", os.path.join(out, "game.html"),
            "-seed", str(args.seed + int(os.path.basename(out).split("-")[1])),
        ] + device_flags(args),
        "self-play",
    )


def device_flags(args):
    """Self-play device settings.

    On the GPU these are not optional extras: a call costs about the same
    whether it carries 50 positions or 950, so throughput is batch size, and
    batch size is how many games are in flight. Measured at 100 sims: 4.8
    games/s on CPU, 1.2 on GPU with 64 games in flight, 18.1 with 1024.
    """
    if not args.cuda:
        return []
    return ["-cuda", "-batch", str(args.batch), "-workers", str(args.workers)]


def train(args, out):
    if os.path.exists(os.path.join(out, "candidate.onnx")):
        log("  training already done, skipping")
        return

    shell(
        [
            sys.executable, os.path.join("training", "train.py"),
            "--run", args.run, "--out", out,
            "--keep", str(args.keep), "--epochs", str(args.epochs),
            "--filters", str(args.filters), "--blocks", str(args.blocks),
        ],
        "training",
    )


def evaluate(args, out):
    """Candidate against the champion, and periodically against the yardstick.

    Self-play rating is purely relative and will climb happily inside a
    delusion, so every --yardstick-every iterations the candidate also plays the
    Voronoi baseline, which never changes and cannot be fooled.
    """
    result_path = os.path.join(out, "eval.json")
    if os.path.exists(result_path):
        log("  evaluation already done, skipping")
        return json.load(open(result_path))

    shell(
        [
            "go", "run", "./cmd/selfplay",
            "-model", os.path.join(out, "candidate.onnx"),
            "-against", os.path.join(args.run, "champion.onnx"),
            "-board", str(args.board),
            "-games", str(args.eval_games),
            "-sims", str(args.sims),
            "-json", result_path,
        ] + device_flags(args),
        "evaluation",
    )

    iteration = int(os.path.basename(out).split("-")[1])
    if args.yardstick_every and iteration % args.yardstick_every == 0:
        shell(
            [
                "go", "run", "./cmd/selfplay",
                "-model", os.path.join(out, "candidate.onnx"),
                "-against", "voronoi",
                "-board", str(args.board),
                "-games", str(args.eval_games),
                "-sims", str(args.sims),
                "-json", os.path.join(out, "yardstick.json"),
                "-html", os.path.join(out, "yardstick.html"),
            ] + device_flags(args),
            "yardstick",
        )

    return json.load(open(result_path))


def promote(args, out):
    for name in ("candidate.onnx", "candidate.pt"):
        source = os.path.join(out, name)
        if os.path.exists(source):
            shutil.copyfile(source, os.path.join(args.run, name.replace("candidate", "champion")))


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--run", default="runs/dev", help="run directory; also the resume point")
    ap.add_argument("--iterations", type=int, default=100, help="total iterations to reach, not to add")
    ap.add_argument("--board", type=int, default=11)
    ap.add_argument("--games", type=int, default=2000, help="self-play games per iteration")
    ap.add_argument("--sims", type=int, default=200, help="search simulations per move")
    ap.add_argument("--eval-games", type=int, default=400)
    ap.add_argument("--keep", type=int, default=10, help="iterations of games in the replay buffer")
    ap.add_argument("--epochs", type=int, default=1)
    ap.add_argument("--filters", type=int, default=64)
    ap.add_argument("--blocks", type=int, default=4)
    ap.add_argument("--seed", type=int, default=0)
    ap.add_argument("--gate", type=float, default=0.0,
                    help="promote when average placement against the champion beats this")
    ap.add_argument("--yardstick-every", type=int, default=5,
                    help="play the Voronoi baseline every N iterations; 0 to never")
    ap.add_argument("--cuda", action="store_true", help="run self-play inference on the GPU")
    ap.add_argument("--batch", type=int, default=4096, help="--cuda: cap on positions per network call")
    ap.add_argument("--workers", type=int, default=1024,
                    help="--cuda: games in flight, which is what sets the batch size")
    args = ap.parse_args()

    if args.cuda and args.workers > args.games:
        # Batch size is games in flight, so more workers than games just leaves
        # the batch short of what the hardware wants.
        log(f"note: --workers {args.workers} exceeds --games {args.games}; batches cannot exceed the game count")

    args.run = os.path.join(ROOT, args.run) if not os.path.isabs(args.run) else args.run
    find_onnxruntime(args.cuda)
    ensure_champion(args)

    start = completed_iterations(args.run)
    if start:
        log(f"resuming at iteration {start} ({start} already finished)")
    if start >= args.iterations:
        log(f"nothing to do: {start} iterations already finished, --iterations is {args.iterations}")
        return

    for i in range(start, args.iterations):
        out = iteration_dir(args.run, i)
        os.makedirs(out, exist_ok=True)
        began = time.time()
        log(f"=== iteration {i}/{args.iterations - 1} ===")

        self_play(args, out)
        train(args, out)
        result = evaluate(args, out)

        placement = result["placement"]
        promoted = placement > args.gate
        if promoted:
            promote(args, out)

        log(f"  placement {placement:+.3f} vs champion, finishes {result['finishes']} "
            f"-> {'PROMOTED' if promoted else 'kept champion'}  ({time.time() - began:.0f}s)")

        with open(os.path.join(args.run, "log.jsonl"), "a") as f:
            f.write(json.dumps({
                "iteration": i,
                "placement": placement,
                "finishes": result["finishes"],
                "promoted": promoted,
                "seconds": round(time.time() - began, 1),
            }) + "\n")


if __name__ == "__main__":
    main()
