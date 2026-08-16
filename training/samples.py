"""Reading the training records self-play writes.

The layout mirrors internal/sim/constrictor/sample.go. Both sides are fixed
size and little-endian, so a whole iteration is one numpy.fromfile.
"""

import glob
import os
import struct

import numpy as np

from encode import PLANES, encode

MAGIC = b"CSAM"
VERSION = 1
MAX_CELLS = 121  # constrictor.MaxCells - fixed, so the record size is board-independent
MAX_SNAKES = 4

HEADER = struct.Struct("<4sBBBB")  # magic, version, w, h, seats

RECORD = np.dtype(
    [
        ("turn", "<i2"),
        ("ego", "u1"),
        ("alive", "u1"),
        ("cells", "u1", MAX_CELLS),
        ("heads", "u1", MAX_SNAKES),
        ("policy", "<f4", 4),
        ("value", "<f4"),
    ]
)
assert RECORD.itemsize == 149, f"record layout drifted from Go: {RECORD.itemsize} bytes"


def read(path):
    """Returns (records, w, h, seats). Records are the RECORD dtype."""
    with open(path, "rb") as f:
        magic, version, w, h, seats = HEADER.unpack(f.read(HEADER.size))
        if magic != MAGIC:
            raise ValueError(f"{path}: not a sample file")
        if version != VERSION:
            raise ValueError(f"{path}: version {version}, expected {VERSION} - regenerate or migrate")
        records = np.fromfile(f, dtype=RECORD)
    return records, w, h, seats


def read_iterations(run_dir, keep):
    """Reads the replay buffer: the newest `keep` iterations of games.

    Older iterations are dropped rather than kept forever, because a buffer
    stretching back to the random-play era teaches the network to predict moves
    nobody would make now.
    """
    paths = sorted(glob.glob(os.path.join(run_dir, "iter-*", "games.bin")))[-keep:]
    if not paths:
        raise FileNotFoundError(f"no games.bin under {run_dir}")

    chunks, meta = [], None
    for path in paths:
        records, w, h, seats = read(path)
        if meta and meta != (w, h, seats):
            raise ValueError(f"{path}: board {w}x{h}/{seats} does not match {meta} earlier in the buffer")
        meta = (w, h, seats)
        chunks.append(records)

    return np.concatenate(chunks), meta, [os.path.basename(os.path.dirname(p)) for p in paths]


def to_planes(records, w, h):
    """Encodes records into network inputs, one row each."""
    out = np.zeros((len(records), PLANES, h, w), dtype=np.float32)
    for i, r in enumerate(records):
        out[i] = encode(r["cells"], r["heads"], int(r["alive"]), int(r["turn"]), w, h, int(r["ego"]))
    return out
