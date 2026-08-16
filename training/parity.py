"""Phase 1 gate: the same position must produce the same policy and value in
PyTorch and in Go, to 1e-4.

Two ways that can break, checked separately so a failure says which:

  1. The encoder. Go and Python each build the input planes from a raw state,
     and a disagreement here trains the network on a game it never plays.
  2. The export. PyTorch and ONNX Runtime must agree on the same weights.

Run in order:

    go test ./internal/sim/constrictor -run WriteParityFixture   # positions
    .venv/Scripts/python training/parity.py                      # 1 and 2, writes expected outputs
    go test ./internal/sim/nn -run Parity                        # 3: Go's ORT against those outputs

Stage 3 lives in Go because that is the path that actually serves moves.
"""

import json
import os
import sys

import numpy as np
import torch

from encode import encode_sample
from net import Net, export

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
FIXTURE = os.path.join(ROOT, "internal", "sim", "constrictor", "testdata", "encode_parity.json")
TESTDATA = os.path.join(ROOT, "training", "testdata")
ONNX = os.path.join(TESTDATA, "parity.onnx")
EXPECTED = os.path.join(TESTDATA, "parity_expected.json")

TOLERANCE = 1e-4


def check_encoder(samples):
    """Python's encoder against the planes Go wrote for the same positions."""
    worst = 0.0
    for i, s in enumerate(samples):
        want = np.asarray(s["encoded"], dtype=np.float32).reshape(-1)
        got = encode_sample(s).reshape(-1)
        if got.shape != want.shape:
            sys.exit(f"sample {i}: Go encoded {want.shape[0]} values, Python {got.shape[0]}")
        worst = max(worst, float(np.abs(got - want).max()))

    if worst > TOLERANCE:
        sys.exit(f"FAIL encoder: worst disagreement {worst:.2e} over {len(samples)} positions")
    print(f"  encoder      Go vs Python, {len(samples)} positions, worst {worst:.2e}")


def check_export(net, inputs):
    """PyTorch against ONNX Runtime on the exported weights."""
    try:
        import onnxruntime as ort
    except ImportError:
        sys.exit("onnxruntime not installed: pip install -r training/requirements.txt")

    with torch.no_grad():
        policy, value = net(torch.from_numpy(inputs))
    policy, value = policy.numpy(), value.numpy()

    session = ort.InferenceSession(ONNX, providers=["CPUExecutionProvider"])
    ort_policy, ort_value = session.run(None, {"board": inputs})

    worst = max(float(np.abs(policy - ort_policy).max()), float(np.abs(value - ort_value).max()))
    if worst > TOLERANCE:
        sys.exit(f"FAIL export: PyTorch and ONNX Runtime differ by {worst:.2e}")
    print(f"  export       PyTorch vs ONNX Runtime, worst {worst:.2e}")
    return policy, value


def main():
    if not os.path.exists(FIXTURE):
        sys.exit(f"no fixture at {FIXTURE}\nrun: go test ./internal/sim/constrictor -run WriteParityFixture")

    with open(FIXTURE) as f:
        samples = json.load(f)["samples"]
    board = samples[0]["w"]

    print(f"parity over {len(samples)} positions, {board}x{board}, tolerance {TOLERANCE:g}")
    check_encoder(samples)

    # Untrained weights are fine and the seed is what makes them reproducible:
    # the gate is about two implementations agreeing, not about playing well.
    torch.manual_seed(0)
    net = Net(board=board).eval()
    os.makedirs(TESTDATA, exist_ok=True)
    export(net, ONNX, board)

    inputs = np.stack([encode_sample(s) for s in samples]).astype(np.float32)
    policy, value = check_export(net, inputs)

    with open(EXPECTED, "w") as f:
        json.dump({"policy": policy.tolist(), "value": value.tolist()}, f)
    print(f"  expected outputs -> {os.path.relpath(EXPECTED, ROOT)}")
    print("\nnow run: go test ./internal/sim/constrictor -run Parity")


if __name__ == "__main__":
    main()
