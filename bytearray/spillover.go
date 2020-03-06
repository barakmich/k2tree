package bytearray

import (
	"fmt"
	"math"

	"github.com/tmthrgd/go-popcount"
)

type SpilloverArray struct {
	bytes      []byte
	levelOff   []int
	length     int
	pagesize   int
	highwater  int
	low        int
	multiplier bool
}

func (a *SpilloverArray) stats() string {
	var b SpilloverArray
	b = *a
	b.bytes = nil
	return fmt.Sprintf("%#v", b)
}

func NewSpillover(pagesize int, highwaterPercentage, lowUtilization float64, multiplier bool) *SpilloverArray {
	if highwaterPercentage < lowUtilization {
		panic("User error: highwaterPercentage is higher than lowUtilization")
	}
	hw := int(math.Round(highwaterPercentage * float64(pagesize)))
	low := int(math.Round(lowUtilization * float64(pagesize)))

	return &SpilloverArray{
		bytes:      make([]byte, pagesize),
		levelOff:   []int{pagesize},
		length:     0,
		pagesize:   pagesize,
		highwater:  hw,
		low:        low,
		multiplier: multiplier,
	}
}

func (a *SpilloverArray) Set(idx int, b byte) {
	if idx >= a.length {
		panic("Writing off the edge of the Spillover Array")
	}
	_, off := a.findOffset(idx)
	a.bytes[off] = b
}

func (a *SpilloverArray) Get(idx int) byte {
	if idx >= a.length {
		panic("Fetching off the edge of the Spillover Array")
	}
	_, off := a.findOffset(idx)
	return a.bytes[off]
}

func (a *SpilloverArray) Insert(idx int, b []byte) {
	for len(b) != 0 {
		l, absoff := a.findOffset(idx)
		var toInsert []byte
		free := a.levelFree(l)
		if len(b) <= free {
			toInsert = b
			b = nil
		} else {
			toInsert = b[:free]
			b = b[free:]
		}
		a.insertIntoLevel(l, absoff, toInsert)
		if a.needsBalance(l) {
			a.rebalance()
		}
	}
}

func (a *SpilloverArray) levelPower(n int, l int) int {
	if a.multiplier {
		return n << l
	}
	return n
}

func (a *SpilloverArray) levelCount(l int) int {
	return a.levelStart(l+1) - a.levelOff[l]
}

func (a *SpilloverArray) levelStart(l int) int {
	if a.multiplier {
		return (a.pagesize << l) - a.pagesize
	}
	return a.pagesize * l
}

func (a *SpilloverArray) insertIntoLevel(level int, absindex int, b []byte) {
	off := a.levelOff[level]
	copy(a.bytes[off-len(b):], a.bytes[off:absindex])
	a.levelOff[level] -= len(b)
	a.length += len(b)
	copy(a.bytes[absindex-len(b):], b)
}

func (a *SpilloverArray) rebalance() {
	for l := 0; l < a.levels(); l++ {
		if a.needsBalance(l) {
			// Time to spill downward
			if l == a.levels()-1 {
				a.createNewLevel()
			}
			overlow := a.levelCount(l) - a.levelPower(a.low, l)
			if overlow < 0 {
				panic(fmt.Sprintf("l: %d, off %#v", l, a.levelOff))
			}
			toMove := min(overlow, a.levelFree(l+1))
			a.levelOff[l+1] -= toMove
			copy(a.bytes[a.levelOff[l+1]:], a.bytes[a.levelStart(l+1)-toMove:a.levelStart(l+1)])
			copy(a.bytes[a.levelOff[l]+toMove:a.levelStart(l+1)], a.bytes[a.levelOff[l]:])
			a.levelOff[l] += toMove
		}
	}
}

func (a *SpilloverArray) createNewLevel() {
	newLevel := a.levels()
	a.bytes = append(a.bytes, make([]byte, a.levelTotalCapacity(newLevel))...)
	a.levelOff = append(a.levelOff, len(a.bytes))
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func (a *SpilloverArray) needsBalance(l int) bool {
	return a.levelCount(l) > a.levelPower(a.highwater, l)
}

func (a *SpilloverArray) Len() int {
	return a.length
}

// findOffset finds the real offset in Array.bytes
// that corresponds with the abstracted offset
// as well as the level at which that offset occurs
func (a *SpilloverArray) findOffset(idx int) (level, offset int) {
	var i, x int
	for i, x = range a.levelOff {
		count := a.levelCount(i)
		if idx == 0 {
			return i, x
		}
		if count > idx {
			return i, x + idx
		}
		idx -= count
	}
	if idx == 0 {
		return i, a.levelOff[i] + a.levelCount(i)
	}
	panic(fmt.Sprintf("offset too large %d", idx))
}

func (a *SpilloverArray) levelTotalCapacity(l int) int {
	return a.levelPower(a.pagesize, l)
}

func (a *SpilloverArray) levelUsage(l int) int {
	return a.levelCount(l)
}

func (a *SpilloverArray) levelFree(l int) int {
	return a.levelTotalCapacity(l) - a.levelCount(l)
}

func (a *SpilloverArray) levels() int {
	return len(a.levelOff)
}

func (a *SpilloverArray) checkInvariants() {
	s := 0
	for i := range a.levelOff {
		s += a.levelCount(i)
	}
	if s != a.length {
		panic("length invariant broken")
	}
	s = 0
	for l := 0; l < a.levels(); l++ {
		s += a.levelUsage(l)
	}
	if s != a.length {
		panic("levelUsage invariant broken")
	}

}

func (a *SpilloverArray) PopCount(start, end int) uint64 {
	var count uint64
	startl, startoff := a.findOffset(start)
	endl, endoff := a.findOffset(end)

	if startl == endl {
		return popcount.CountBytes(a.bytes[startoff:endoff])
	}

	count += popcount.CountBytes(a.bytes[startoff:a.levelStart(startl+1)])
	count += popcount.CountBytes(a.bytes[a.levelOff[endl]:endoff])
	for l := startl + 1; l < endl; l++ {
		count += popcount.CountBytes(a.bytes[a.levelOff[l]:a.levelStart(l+1)])
	}
	return count
}
