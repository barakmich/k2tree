package k2tree

import (
	"bytes"
	"testing"
)

func TestSliceArrayInsertTable(t *testing.T) {
	tt := []struct {
		n      int
		at     int
		input  []byte
		output []byte
		length int
	}{
		{
			n:      4,
			at:     12,
			input:  []byte{0xAB, 0xCD, 0xEF},
			length: 24,
			output: []byte{0xAB, 0xC0, 0xDE, 0xF0},
		},
		{
			n:      12,
			at:     4,
			input:  []byte{0xAB, 0xCD, 0xEF},
			length: 24,
			output: []byte{0xA0, 0x00, 0xBC, 0xDE, 0xF0},
		},
		{
			n:      8,
			at:     4,
			input:  []byte{0xAB, 0xCD, 0xEF},
			length: 24,
			output: []byte{0xA0, 0x0B, 0xCD, 0xEF},
		},
		{
			n:      16,
			at:     8,
			input:  []byte{0xAB, 0xCD, 0xEF},
			length: 24,
			output: []byte{0xAB, 0x00, 0x00, 0xCD, 0xEF},
		},
		{
			n:      12,
			at:     8,
			input:  []byte{0xAB, 0xCD, 0xEF},
			length: 24,
			output: []byte{0xAB, 0x00, 0x0C, 0xDE, 0xF0},
		},
		{
			n:      4,
			at:     8,
			input:  []byte{0xAB, 0xCD, 0xEF},
			length: 24,
			output: []byte{0xAB, 0x0C, 0xDE, 0xF0},
		},
		{
			n:      4,
			at:     12,
			input:  []byte{0xAB, 0xCD, 0xEF, 0x10},
			length: 28,
			output: []byte{0xAB, 0xC0, 0xDE, 0xF1},
		},
		{
			n:      4,
			at:     16,
			input:  []byte{0xAB, 0xCD, 0xEF, 0x10},
			length: 28,
			output: []byte{0xAB, 0xCD, 0x0E, 0xF1},
		},
		{
			n:      12,
			at:     8,
			input:  []byte{0xAB, 0xCD, 0xEF, 0x10},
			length: 28,
			output: []byte{0xAB, 0x00, 0x0C, 0xDE, 0xF1},
		},
		{
			n:      12,
			at:     12,
			input:  []byte{0xAB, 0xCD, 0xEF, 0x10},
			length: 28,
			output: []byte{0xAB, 0xC0, 0x00, 0xDE, 0xF1},
		},
		{
			n:      4,
			at:     4,
			input:  []byte{0x19},
			length: 8,
			output: []byte{0x10, 0x90},
		},
	}
	for _, x := range tt {
		s := &sliceArray{
			bytes:  x.input,
			length: x.length,
		}
		s.Insert(x.n, x.at)
		if !bytes.Equal(s.bytes, x.output) {
			t.Errorf("mismatch! got %#v expected %#v (n: %d, at %d, len %d)", s.bytes, x.output, x.n, x.at, x.length)
		}
	}
}
