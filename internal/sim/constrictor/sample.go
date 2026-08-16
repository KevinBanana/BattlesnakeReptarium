package constrictor

// sample.go is the training record: what self-play hands to the trainer.
//
// One record per living seat per turn, fixed size and little-endian, so Python
// reads a whole iteration with a single numpy.fromfile and a structured dtype.
// training/samples.py holds the matching definition - the two are checked
// against each other by the sample round-trip test.
//
// Positions are stored raw rather than as encoded planes: a State is 128 bytes
// where its five planes are 2.4KB, and an iteration is on the order of half a
// million samples. Python builds the planes with the same code the trainer
// already needs.

import (
	"bufio"
	"encoding/binary"
	"io"
)

// SampleMagic identifies the format, and the version rises whenever the record
// layout changes so an old file fails loudly instead of being misread.
var SampleMagic = [4]byte{'C', 'S', 'A', 'M'}

const SampleVersion = 1

// SampleHeader precedes the records. There is no count: the reader consumes
// records until EOF, so a run killed mid-write still yields every completed
// game rather than an unreadable file.
type SampleHeader struct {
	Magic   [4]byte
	Version uint8
	W, H    uint8
	Seats   uint8
}

// Sample is one seat's view of one position, with what search decided and how
// the game turned out. Cells is always MaxCells so the record is fixed size on
// any board; the header says how much of it is real.
type Sample struct {
	Turn   int16
	Ego    uint8
	Alive  uint8
	Cells  [MaxCells]uint8
	Heads  [MaxSnakes]uint8
	Policy [4]float32 // the search's visit distribution: the policy target
	Value  float32    // that seat's final placement: the value target
}

// NewSample captures the position and the search result. Value is left zero
// until the game ends and the placement is known.
func NewSample(s *State, ego int, policy [4]float32) Sample {
	return Sample{
		Turn:   int16(s.Turn),
		Ego:    uint8(ego),
		Alive:  s.Alive,
		Cells:  s.Cells,
		Heads:  s.Heads,
		Policy: policy,
	}
}

// SampleWriter appends records to a file.
type SampleWriter struct {
	w *bufio.Writer
}

func NewSampleWriter(out io.Writer, w, h, seats int) (*SampleWriter, error) {
	buf := bufio.NewWriterSize(out, 1<<20)
	header := SampleHeader{Magic: SampleMagic, Version: SampleVersion, W: uint8(w), H: uint8(h), Seats: uint8(seats)}
	if err := binary.Write(buf, binary.LittleEndian, header); err != nil {
		return nil, err
	}
	return &SampleWriter{w: buf}, nil
}

func (sw *SampleWriter) Write(samples []Sample) error {
	return binary.Write(sw.w, binary.LittleEndian, samples)
}

func (sw *SampleWriter) Flush() error { return sw.w.Flush() }
