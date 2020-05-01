package k2tree

import (
	"bytes"
	"testing"
)

func TestInsertFourBits(t *testing.T) {
	tt := []struct {
		in          []byte
		inByte      byte
		expected    []byte
		expectedOut byte
	}{
		{
			in:          []byte{0x01, 0x02, 0x03},
			inByte:      0x0F,
			expected:    []byte{0xF0, 0x10, 0x20},
			expectedOut: 0x30,
		},
		{
			in:          []byte{0x1A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A},
			inByte:      0x0F,
			expected:    []byte{0xF1, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0},
			expectedOut: 0xA0,
		},
	}
	for i, x := range tt {
		b := make([]byte, len(x.in))
		copy(b, x.in)
		gotOut := insertFourBits(b, x.inByte)
		if !bytes.Equal(b, x.expected) {
			t.Logf("Bytes not matching for test %d: got %#v, expected %#v", i, b, x.expected)
			t.Fail()
		}
		if gotOut != x.expectedOut {
			t.Logf("Output byte not matching for test %d: got %#v, expected %#v", i, gotOut, x.expectedOut)
			t.Fail()
		}
	}
}
