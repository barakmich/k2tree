package k2tree

import (
	"bytes"
	"math/rand"
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
			inByte:      0xF0,
			expected:    []byte{0xF0, 0x10, 0x20},
			expectedOut: 0x30,
		},
		{
			in:          []byte{0x1A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A},
			inByte:      0xF0,
			expected:    []byte{0xF1, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0},
			expectedOut: 0xA0,
		},
		{
			in: []byte{0x1A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A,
				0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A, 0x0A},
			inByte: 0xF0,
			expected: []byte{0xF1, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0,
				0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0, 0xA0},
			expectedOut: 0xA0,
		},
		{
			in:          []byte{0xef, 0x4, 0xf6, 0x14, 0x4f, 0x21, 0x78, 0x5a, 0x55, 0x66, 0xa1, 0x3d, 0xf8, 0x7d, 0x5c, 0x10, 0x3c, 0x3f, 0x28, 0x4c, 0x5, 0x62, 0x3f, 0xda, 0xbf, 0x11, 0xa7, 0x2, 0x8e, 0xc8, 0x4f, 0xc2},
			inByte:      0xa7,
			expected:    []byte{0xae, 0xf0, 0x4f, 0x61, 0x44, 0xf2, 0x17, 0x85, 0xa5, 0x56, 0x6a, 0x13, 0xdf, 0x87, 0xd5, 0xc1, 0x3, 0xc3, 0xf2, 0x84, 0xc0, 0x56, 0x23, 0xfd, 0xab, 0xf1, 0x1a, 0x70, 0x28, 0xec, 0x84, 0xfc},
			expectedOut: 0x20,
		},
	}
	for i, x := range tt {
		b := make([]byte, len(x.in))
		copy(b, x.in)
		gotOut := insertFourBits(b, x.inByte)
		if !bytes.Equal(b, x.expected) {
			t.Logf("Bytes not matching for test %d: \ngot\t\t %#v\nexpected\t %#v", i, b, x.expected)
			t.Fail()
		}
		if gotOut != x.expectedOut {
			t.Logf("Output byte not matching for test %d: got %#v, expected %#v", i, gotOut, x.expectedOut)
			t.Fail()
		}
	}
}

const FUZZ_ITER = 3000

func TestInsertFourFuzz(t *testing.T) {
	for i := 0; i < FUZZ_ITER; i++ {
		inbyte := byte(rand.Intn(256))
		orig := generateByteString((rand.Intn(256) * rand.Intn(256)) + 1)
		bytestr := make([]byte, len(orig))
		realbyte := make([]byte, len(orig))
		copy(realbyte, orig)
		copy(bytestr, orig)
		out := insertFourBits(bytestr[1:], inbyte)
		outreal := insertFourBitsGo(realbyte[1:], inbyte)
		if !bytes.Equal(bytestr, realbyte) {
			t.Logf("Mismatched test case:\n{\nin: %#v,\ninByte: %#v,\ngot: %#v,\ngotOut: %#v,\nexpected: %#v,\nexpectedOut: %#v,\n},\n",
				orig, inbyte, bytestr, out, realbyte, outreal)
			t.Fail()
		} else if out != outreal {
			t.Logf("Mismatched test case:\n{\nin: %#v,\ninByte: %#v,\ngot: %#v,\ngotOut: %#v,\nexpected: %#v,\nexpectedOut: %#v,\n},\n",
				orig, inbyte, bytestr, out, realbyte, outreal)
			t.Fail()
		}
	}
}

func generateByteString(size int) []byte {
	out := make([]byte, size)
	for i := 0; i < size; i++ {
		b := byte(rand.Intn(256))
		out[i] = b
	}
	return out
}

func TestInsertFourMonadic(t *testing.T) {
	// Checks if calling insertFourBits followed by insertFourBits and passing
	// along the extra byte is equivalent.
}
