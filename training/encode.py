"""State -> network input.

A line-for-line mirror of internal/sim/constrictor/encode.go. This is the one
piece of logic that has to exist in both languages: Go encodes at inference
time, Python encodes at training time, and if they disagree the network is
trained on a game it never plays. parity.py is what keeps them honest - run it
after touching either side.
"""

import numpy as np

PLANE_OWN_BODY = 0
PLANE_OWN_HEAD = 1
PLANE_ENEMY_BODIES = 2
PLANE_ENEMY_HEADS = 3
PLANE_TURN = 4
PLANES = 5

EMPTY = 255  # constrictor.Empty


def encode(cells, heads, alive, turn, w, h, ego):
    """Returns seat ego's view as a (PLANES, h, w) float32 array.

    cells is one owning seat index per cell in y*w+x order, EMPTY for free.
    heads is one cell per seat, meaningful only for seats whose bit is set in
    alive - Apply leaves a dead seat's entry stale, exactly as in Go.
    """
    cells = np.asarray(cells, dtype=np.uint8)
    out = np.zeros((PLANES, h * w), dtype=np.float32)

    occupied = cells != EMPTY
    out[PLANE_OWN_BODY][occupied & (cells == ego)] = 1
    out[PLANE_ENEMY_BODIES][occupied & (cells != ego)] = 1

    for seat, head in enumerate(heads):
        if not (alive >> seat) & 1:
            continue
        out[PLANE_OWN_HEAD if seat == ego else PLANE_ENEMY_HEADS][head] = 1

    # Same scale as Go: the longest game the board can hold, since every living
    # snake is length 3+turn and two survivors fill w*h by turn w*h/2 - 3.
    out[PLANE_TURN] = np.float32(turn) / (np.float32(w * h) / 2 - 3)

    return out.reshape(PLANES, h, w)


def encode_sample(s):
    """Encodes one record as written by Go, keyed the same way."""
    return encode(s["cells"], s["heads"], s["alive"], s["turn"], s["w"], s["h"], s["ego"])
