package k2tree

import (
	"fmt"
	"math"
)

type binaryLRUIndex struct {
	bits         bitarray
	offsets      []int
	counts       []int
	cacheHistory []int
	size         int
}

var _ bitarray = (*binaryLRUIndex)(nil)

const PopcntCacheBits = 1024

func newBinaryLRUIndex(bits bitarray, size int) *binaryLRUIndex {
	return &binaryLRUIndex{
		bits: bits,
		size: size,
	}
}

// Len returns the number of bits in the bitarray.
func (b *binaryLRUIndex) Len() int {
	return b.bits.Len()
}

// Set sets the bit at an index `at` to the value `val`.
func (b *binaryLRUIndex) Set(at int, val bool) {
	cur := b.bits.Get(at)
	if cur && val {
		return
	}
	if !cur && !val {
		return
	}
	b.bits.Set(at, val)
	var delta int
	if val {
		delta = 1
	} else {
		delta = -1
	}
	for i, o := range b.offsets {
		if at < o {
			b.counts[i] += delta
		}
	}

}

// Get returns the value stored at `at`.
func (b *binaryLRUIndex) Get(at int) bool {
	return b.bits.Get(at)
}

// Count returns the number of set bits in the interval [from, to).
func (b *binaryLRUIndex) Count(from int, to int) int {
	if from == to {
		return 0
	}
	var subresult int
	result, cacheit := b.zeroCount(to)
	if cacheit {
		b.cacheAdd(result, to)
	}
	if from != 0 {
		subresult, cacheit = b.zeroCount(from)
		if cacheit {
			b.cacheAdd(subresult, from)
		}
		result = result - subresult
	}
	return result
}

func (b *binaryLRUIndex) zeroCount(to int) (int, bool) {
	cacheit := false
	count, at := b.getClosestCache(to)
	if abs(to-at) > PopcntCacheBits {
		cacheit = true

	}
	if at == to {
		return count, cacheit
	} else if at < to {
		return count + b.bits.Count(at, to), cacheit
	} else {
		return count - b.bits.Count(to, at), cacheit
	}
}

func (b *binaryLRUIndex) getClosestCache(to int) (count, at int) {
	if len(b.offsets) == 0 {
		return 0, 0
	}
	idx := bSearch(b.offsets, to)
	downdist := math.MaxInt64
	if idx != 0 {
		downdist = to - b.offsets[idx-1]
	}
	updist := math.MaxInt64
	if idx != len(b.offsets) {
		updist = b.offsets[idx] - to
	}
	if downdist < updist {
		b.cacheHit(idx - 1)
		return b.counts[idx-1], b.offsets[idx-1]
	}
	b.cacheHit(idx)
	return b.counts[idx], b.offsets[idx]
}

func (b *binaryLRUIndex) cacheHit(idx int) {
	for i := len(b.cacheHistory) - 1; i >= 0; i-- {
		if b.cacheHistory[i] == idx {
			cut := append(b.cacheHistory[:i], b.cacheHistory[i+1:]...)
			b.cacheHistory = append(cut, idx)
			return
		}
	}
	panic("idx must be in cacheHistory")
}

func (b *binaryLRUIndex) cacheAdd(val, at int) {
	if len(b.offsets) >= b.size {
		b.cacheEvict()
	}
	idx := bSearch(b.offsets, at)
	b.offsets = append(b.offsets, 0)
	copy(b.offsets[idx+1:], b.offsets[idx:])
	b.offsets[idx] = at
	b.counts = append(b.counts, 0)
	copy(b.counts[idx+1:], b.counts[idx:])
	b.counts[idx] = val
	for i := 0; i < len(b.cacheHistory); i++ {
		if b.cacheHistory[i] >= idx {
			b.cacheHistory[i]++
		}
	}
	assert(len(b.cacheHistory) <= b.size, fmt.Sprint(b.cacheHistory))
	b.cacheHistory = append(b.cacheHistory, idx)
}

func (b *binaryLRUIndex) cacheEvict() {
	//pop
	todel := b.cacheHistory[0]
	b.cacheHistory = b.cacheHistory[1:]
	for i, v := range b.cacheHistory {
		if v >= todel {
			b.cacheHistory[i]--
		}
	}
	b.offsets = append(b.offsets[:todel], b.offsets[todel+1:]...)
	b.counts = append(b.counts[:todel], b.counts[todel+1:]...)
}

// Total returns the total number of set bits.
func (b *binaryLRUIndex) Total() int {
	return b.bits.Total()
}

// Bytes returns the bitarray as a byte array
func (b *binaryLRUIndex) Bytes() []byte {
	return b.bits.Bytes()
}

// Insert extends the bitarray by `n` bits. The bits are zeroed
// and start at index `at`. Example:
// Initial string: 11101
// Insert(3, 2)
// Resulting string: 11000101
func (b *binaryLRUIndex) Insert(n int, at int) error {
	err := b.bits.Insert(n, at)
	if err != nil {
		return err
	}
	for i := 0; i < len(b.offsets); i++ {
		if at < b.offsets[i] {
			b.offsets[i] += n
		}
	}
	return nil
}

func (b *binaryLRUIndex) debug() string {
	return fmt.Sprintf("BinaryLRUIndex:\n internal: %s, %#v", b.bits.debug(), b)
}

func bSearch(arr []int, x int) int {
	n := len(arr)
	// Define f(-1) == false and f(n) == true.
	// Invariant: f(i-1) == false, f(j) == true.
	i, j := 0, n
	for i < j {
		h := int(uint(i+j) >> 1) // avoid overflow when computing h
		// i ≤ h < j
		if arr[h] < x {
			i = h + 1 // preserves f(i-1) == false
		} else {
			j = h // preserves f(j) == true
		}
	}
	// i == j, f(i-1) == false, and f(j) (= f(i)) == true  =>  answer is i.
	return i
}
