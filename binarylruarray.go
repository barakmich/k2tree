package k2tree

import (
	"fmt"
	"math"
)

type binaryLRUIndex struct {
	bits          bitarray
	offsets       []int
	counts        []int
	historyMap    []int
	size          int
	tick          int
	cacheDistance int
}

var _ bitarray = (*binaryLRUIndex)(nil)

const (
	// DefaultLRUCacheDistance was optimized experimentally. It's the distance
	// in bits between cache hits. It's a tradeoff between leaning on the POPCNT
	// instruction between known offsets in the cache and the overhead of
	// maintaining the LRU. If the LRU gets cheaper to maintain, this may get
	// decreased. If POPCNT gets faster, this may increase.
	DefaultLRUCacheDistance = 512
)

func newBinaryLRUIndex(bits bitarray, size int) *binaryLRUIndex {
	return &binaryLRUIndex{
		bits:          bits,
		size:          size,
		cacheDistance: DefaultLRUCacheDistance,
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
	result := b.zeroCount(to)
	if from != 0 {
		subresult = b.zeroCount(from)
		result = result - subresult
	}
	return result
}

func (b *binaryLRUIndex) zeroCount(to int) int {
	count, at, _ := b.getClosestCache(to)
	var val int
	if at == to {
		return count
	} else if at < to {
		val = count + b.bits.Count(at, to)
	} else {
		val = count - b.bits.Count(to, at)
	}

	// Update the cache
	if abs(to-at) > b.cacheDistance {
		// If we're far away, add it to the cache
		b.cacheAdd(val, to)
	}

	return val
}

func (b *binaryLRUIndex) getClosestCache(to int) (count, at, idx int) {
	if len(b.offsets) == 0 {
		return 0, 0, -1
	}
	idx = bSearch(b.offsets, to)
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
		return b.counts[idx-1], b.offsets[idx-1], idx - 1
	}
	b.cacheHit(idx)
	return b.counts[idx], b.offsets[idx], idx
}

func (b *binaryLRUIndex) cacheHit(idx int) {
	b.tick++
	b.historyMap[idx] = b.tick
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
	b.historyMap = append(b.historyMap, 0)
	copy(b.historyMap[idx+1:], b.historyMap[idx:])
	b.tick++
	b.historyMap[idx] = b.tick
}

func (b *binaryLRUIndex) cacheEvict() {
	var timedel int = math.MaxInt64
	todel := -1
	for i, time := range b.historyMap {
		if time < timedel {
			todel = i
			timedel = time
		}
	}
	b.offsets = append(b.offsets[:todel], b.offsets[todel+1:]...)
	b.counts = append(b.counts[:todel], b.counts[todel+1:]...)
	b.historyMap = append(b.historyMap[:todel], b.historyMap[todel+1:]...)
}

// Total returns the total number of set bits.
func (b *binaryLRUIndex) Total() int {
	return b.bits.Total()
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
