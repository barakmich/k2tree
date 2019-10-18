package k2tree

import "fmt"

type quartileIndex struct {
	bits    bitarray
	offsets [3]int
	counts  [3]int
}

var _ bitarray = (*quartileIndex)(nil)

func newQuartileIndex(bits bitarray) *quartileIndex {
	q := &quartileIndex{
		bits: bits,
		offsets: [3]int{
			bits.Len() / 4,
			bits.Len() / 2,
			(bits.Len()/2 + bits.Len()/4),
		},
	}
	q.counts[0] = bits.Count(0, q.offsets[0])
	q.counts[1] = bits.Count(q.offsets[0], q.offsets[1]) + q.counts[1]
	q.counts[2] = bits.Count(q.offsets[0], q.offsets[2]) + q.counts[2]
	return q
}

// Len returns the number of bits in the bitarray.
func (q *quartileIndex) Len() int {
	return q.bits.Len()
}

// Set sets the bit at an index `at` to the value `val`.
func (q *quartileIndex) Set(at int, val bool) {
	cur := q.bits.Get(at)
	if cur && val {
		return
	}
	if !cur && !val {
		return
	}
	q.bits.Set(at, val)
	var delta int
	if val {
		delta = 1
	} else {
		delta = -1
	}
	for i, o := range q.offsets {
		if at < o {
			q.counts[i] += delta
		}
	}
}

// Get returns the value stored at `at`.
func (q *quartileIndex) Get(at int) bool {
	return q.bits.Get(at)
}

// Count returns the number of set bits in the interval [from, to).
func (q *quartileIndex) Count(from int, to int) int {
	if from == 0 {
		return q.zeroCount(to)
	}
	return q.zeroCount(to) - q.zeroCount(from)
}

// zeroCount computes the count from zero to the given value.
func (q *quartileIndex) zeroCount(to int) int {
	prevoff := 0
	prevcount := 0
	for i, off := range q.offsets {
		if to < off {
			if off-to < to-prevoff {
				return q.counts[i] - q.bits.Count(to, off)
			} else {
				return q.bits.Count(prevoff, to) + prevcount
			}
		}
		prevoff = off
		prevcount = q.counts[i]
	}
	if q.bits.Len()-to < to-prevoff {
		return q.bits.Total() - q.bits.Count(to, q.bits.Len())
	} else {
		return q.bits.Count(prevoff, to) + prevcount
	}
}

// Total returns the total number of set bits.
func (q *quartileIndex) Total() int {
	return q.bits.Total()
}

// Insert extends the bitarray by `n` bits. The bits are zeroed
// and start at index `at`. Example:
// Initial string: 11101
// Insert(3, 2)
// Resulting string: 11000101
func (q *quartileIndex) Insert(n int, at int) error {
	if n%4 != 0 {
		panic("can only extend by nibbles (multiples of 4)")
	}
	err := q.bits.Insert(n, at)
	if err != nil {
		return err
	}
	newlen := q.bits.Len()
	for i := 0; i < 3; i++ {
		q.adjust(i, n, at, (newlen * (i + 1) / 4))
	}
	return nil
}

func (q *quartileIndex) adjust(i, n, at, newi int) {
	oldi := q.offsets[i]

	assert(newi >= oldi, "Inserting shrunk the array?")

	q.offsets[i] = newi
	if (n + at) < oldi {
		// Entire span below me, adjust for loss.
		q.counts[i] -= q.bits.Count(newi, oldi+n)
	} else if at >= oldi {
		// Entire span above me, adjust for gain.
		q.counts[i] += q.bits.Count(oldi, newi)
	} else {
		// Span intersects me.
		// Stupid answer:
		q.counts[i] = q.bits.Count(0, newi)
	}
}

func (q *quartileIndex) debug() string {
	return fmt.Sprintf("Quartile:\n internal: %s, %#v", q.bits.debug(), q)
}
