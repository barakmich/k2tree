package k2tree

import "fmt"

type int16index struct {
	bits   bitarray
	counts []uint16
}

var _ bitarray = (*int16index)(nil)

const int16Max = 1 << 16

func newInt16Index(b bitarray) *int16index {
	if b.Len() != 0 {
		panic("unimplemented")
	}
	return &int16index{
		bits:   b,
		counts: make([]uint16, 1),
	}
}

// Len returns the number of bits in the bitarray.
func (ix *int16index) Len() int {
	return ix.bits.Len()
}

// Set sets the bit at an index `at` to the value `val`.
func (ix *int16index) Set(at int, val bool) {
	cur := ix.bits.Get(at)
	if cur && val {
		return
	}
	if !cur && !val {
		return
	}
	ix.bits.Set(at, val)
	var delta int
	if val {
		delta = 1
	} else {
		delta = -1
	}
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
func (ix *int16index) Get(at int) bool {
	return ix.bits.Get(at)
}

// Count returns the number of set bits in the interval [from, to).
func (ix *int16index) Count(from int, to int) int {
	if from == 0 {
		return ix.zeroCount(to)
	}
	out := ix.zeroCount(to) - ix.zeroCount(from)
	return out
}

func (ix *int16index) zeroCount(to int) int {
	// There's a speedup here where we're adding values only to subtract them.
	total := 0
	i := 0
	ioff := 0
	for i = 0; i < len(ix.counts); i++ {
		ioff = (i + 1) * int16Max
		if ioff >= to {
			break
		}
		total += int(ix.counts[i])
	}
	offset := to - (i * int16Max)
	if offset > (int16Max / 2) {
		total += int(ix.counts[i])
		total -= ix.bits.Count(to, min(ioff, ix.bits.Len()))
	} else {
		total += ix.bits.Count(i*int16Max, to)
	}
	//assert(total == ix.bits.Count(0, to), "Debug: Counts don't match")
	return total
}

// Total returns the total number of set bits.
func (ix *int16index) Total() int {
	return ix.bits.Total()
}

// Bytes returns the bitarray as a byte array
func (ix *int16index) Bytes() []byte {
	return ix.bits.Bytes()
}

// Insert extends the bitarray by `n` bits. The bits are zeroed
// and start at index `at`. Example:
// Initial string: 11101
// Insert(3, 2)
// Resulting string: 11000101
func (ix *int16index) Insert(n int, at int) error {
	entries := (ix.bits.Len() + n) / int16Max
	for len(ix.counts) < entries+1 {
		// extend
		ix.counts = append(ix.counts, 0)
	}
	err := ix.bits.Insert(n, at)
	if err != nil {
		return err
	}
	if n >= int16Max {
		ix.adjustBig(at)
	} else {
		ix.adjust(n, at)
	}
	//	ix.adjustBig(at)
	return nil
}

func (ix *int16index) adjustBig(at int) {
	for i := range ix.counts {
		off := (i + 1) * int16Max
		if at >= off {
			continue
		} else {
			c := ix.bits.Count(off-int16Max, min(off, ix.bits.Len()))
			ix.counts[i] = uint16(c)
		}
	}
}

func (ix *int16index) adjust(n, at int) {
	bitlen := ix.bits.Len()
	for i := range ix.counts {
		off := (i + 1) * int16Max
		if at >= off {
			continue
		} else if (at + n) < off-int16Max {
			if off >= bitlen {
				// Do the last one the easy way
				c := ix.bits.Count(off-int16Max, bitlen)
				ix.counts[i] = uint16(c)
				return
			}
			del := ix.bits.Count(off, off+n)
			add := ix.bits.Count(off-int16Max, off-int16Max+n)
			ix.counts[i] = uint16(int(ix.counts[i]) + add - del)
		} else {
			c := ix.bits.Count(off-int16Max, min(off, ix.bits.Len()))
			ix.counts[i] = uint16(c)
		}
	}
}

func (ix *int16index) debug() string {
	return fmt.Sprintf("Int16Index:\n internal: %s\nindex:%#v", ix.bits.debug(), ix.counts)
}
