package bytearray

import (
	"math/bits"
)

type Int16Index struct {
	bytes  ByteArray
	counts []uint16
}

var _ ByteArray = (*Int16Index)(nil)

const int16Max = 1 << 13

func NewInt16Index(b ByteArray) *Int16Index {
	if b.Len() != 0 {
		panic("unimplemented")
	}
	return &Int16Index{
		bytes:  b,
		counts: make([]uint16, 1),
	}
}

// Len returns the number of bytes in the bytearray.
func (ix *Int16Index) Len() int {
	return ix.bytes.Len()
}

// Set sets the bit at an index `at` to the value `val`.
func (ix *Int16Index) Set(at int, val byte) {
	cur := ix.bytes.Get(at)
	ix.bytes.Set(at, val)
	delta := bits.OnesCount8(val) - bits.OnesCount8(cur)
	prevoff := 0
	for i := range ix.counts {
		off := (i + 1) * int16Max
		if at < off && at >= prevoff {
			ix.counts[i] = uint16(int(ix.counts[i]) + delta)
			break
		}
		prevoff = off
	}
}

// Get returns the value stored at `at`.
func (ix *Int16Index) Get(at int) byte {
	return ix.bytes.Get(at)
}

// Count returns the number of set bits in the interval [from, to).
func (ix *Int16Index) PopCount(from int, to int) uint64 {
	if from == 0 {
		return ix.zeroCount(to)
	}
	out := ix.zeroCount(to) - ix.zeroCount(from)
	return out
}

func (ix *Int16Index) zeroCount(to int) uint64 {
	// There's a speedup here where we're adding values only to subtract them.
	var total uint64
	i := 0
	ioff := 0
	for i = 0; i < len(ix.counts); i++ {
		ioff = (i + 1) * int16Max
		if ioff >= to {
			break
		}
		total += uint64(ix.counts[i])
	}
	offset := to - (i * int16Max)
	if offset > (int16Max / 2) {
		total += uint64(ix.counts[i])
		total -= ix.bytes.PopCount(to, min(ioff, ix.bytes.Len()))
	} else {
		total += ix.bytes.PopCount(i*int16Max, to)
	}
	//assert(total == ix.bits.Count(0, to), "Debug: Counts don't match")
	return total
}

// Insert extends the bitarray by `n` bits. The bits are zeroed
// and start at index `at`. Example:
// Initial string: 11101
// Insert(3, 2)
// Resulting string: 11000101
func (ix *Int16Index) Insert(at int, b []byte) {
	n := len(b)
	entries := (ix.bytes.Len() + n) / int16Max
	for len(ix.counts) < entries+1 {
		// extend
		ix.counts = append(ix.counts, 0)
	}
	ix.bytes.Insert(at, make([]byte, n))
	for i := 0; i < n; i++ {
		ix.bytes.Set(at+i, b[i])
	}
	if n >= int16Max {
		ix.adjustBig(at)
	} else {
		ix.adjust(n, at)
	}
}

func (ix *Int16Index) adjustBig(at int) {
	for i := range ix.counts {
		off := (i + 1) * int16Max
		if at >= off {
			continue
		} else {
			c := ix.bytes.PopCount(off-int16Max, min(off, ix.bytes.Len()))
			ix.counts[i] = uint16(c)
		}
	}
}

func (ix *Int16Index) adjust(n, at int) {
	bitlen := ix.bytes.Len()
	for i := range ix.counts {
		off := (i + 1) * int16Max
		if at >= off {
			continue
		} else if (at + n) < off-int16Max {
			if off >= bitlen {
				// Do the last one the easy way
				c := ix.bytes.PopCount(off-int16Max, bitlen)
				ix.counts[i] = uint16(c)
				return
			}
			del := ix.bytes.PopCount(off, off+n)
			add := ix.bytes.PopCount(off-int16Max, off-int16Max+n)
			ix.counts[i] = uint16(uint64(ix.counts[i]) + add - del)
		} else {
			c := ix.bytes.PopCount(off-int16Max, min(off, ix.bytes.Len()))
			ix.counts[i] = uint16(c)
		}
	}
}

func (ix *Int16Index) Copy(from, to, n int) {
	ix.bytes.Copy(from, to, n)
	// Completely recalculate starting at "to"
	ix.adjustBig(to)
}
