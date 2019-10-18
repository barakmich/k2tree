package k2tree

import (
	"container/list"
	"fmt"
	"math"
)

type binaryLRUIndex struct {
	bits         bitarray
	offsets      []int
	counts       []int
	cacheHistory *list.List
	size         int
}

var _ bitarray = (*binaryLRUIndex)(nil)

const (
	PopcntMoveBits  = 10
	PopcntCacheBits = 1024
	PopcntMoveAfter = 512
)

func newBinaryLRUIndex(bits bitarray, size int) *binaryLRUIndex {
	return &binaryLRUIndex{
		bits:         bits,
		size:         size,
		cacheHistory: list.New(),
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
	count, at, idx := b.getClosestCache(to)
	var val int
	if at == to {
		return count
	} else if at < to {
		val = count + b.bits.Count(at, to)
	} else {
		val = count - b.bits.Count(to, at)
	}

	// Update the cache
	if abs(to-at) > PopcntCacheBits {
		// If we're far away, add it to the cache
		b.cacheAdd(val, to)
	} else if abs(to-at) < PopcntMoveBits && to > PopcntMoveAfter {
		// Move the value in the cache to the one just computed.
		b.cacheMove(idx, to, val)
	}

	return val
}

func (b *binaryLRUIndex) cacheMove(idx, to, val int) {
	if idx < 0 {
		return
	}
	b.counts[idx] = val
	b.offsets[idx] = to
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
	for e := b.cacheHistory.Front(); e != nil; e = e.Next() {
		if e.Value.(int) == idx {
			b.cacheHistory.MoveToFront(e)
			return
		}
	}
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
	for e := b.cacheHistory.Front(); e != nil; e = e.Next() {
		t := e.Value.(int)
		if e.Value.(int) >= idx {
			t++
			e.Value = t
		}
	}
	b.cacheHistory.PushFront(idx)
}

func (b *binaryLRUIndex) cacheEvict() {
	//pop
	lastelem := b.cacheHistory.Back()
	todel := lastelem.Value.(int)
	b.cacheHistory.Remove(lastelem)
	for e := b.cacheHistory.Front(); e != nil; e = e.Next() {
		t := e.Value.(int)
		if t >= todel {
			t--
			e.Value = t
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
