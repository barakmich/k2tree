package spillover

import (
	"fmt"
	"math"
)

type Array struct {
	bytes      []byte
	levelOff   []int
	levelCount []int
	length     int
	pagesize   int
	highwater  int
	low        int
}

func (a *Array) stats() string {
	var b Array
	b = *a
	b.bytes = nil
	return fmt.Sprintf("%#v", b)
}

func New(pagesize int, highwaterPercentage, lowUtilization float64) *Array {
	if highwaterPercentage < lowUtilization {
		panic("User error: highwaterPercentage is higher than lowUtilization")
	}
	hw := int(math.Round(highwaterPercentage * float64(pagesize)))
	low := int(math.Round(lowUtilization * float64(pagesize)))

	return &Array{
		bytes:      make([]byte, pagesize),
		levelOff:   []int{pagesize},
		levelCount: []int{0},
		length:     0,
		pagesize:   pagesize,
		highwater:  hw,
		low:        low,
	}
}

func (a *Array) Set(idx int, b byte) {
	if idx >= a.length {
		panic("Writing off the edge of the Spillover Array")
	}
	_, off := a.findOffset(idx)
	a.bytes[off] = b
}

func (a *Array) Get(idx int) byte {
	if idx >= a.length {
		panic("Fetching off the edge of the Spillover Array")
	}
	_, off := a.findOffset(idx)
	return a.bytes[off]
}

func (a *Array) Insert(idx int, b []byte) {
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

func (a *Array) insertIntoLevel(level int, absindex int, b []byte) {
	off := a.levelOff[level]
	copy(a.bytes[off-len(b):], a.bytes[off:absindex])
	a.levelOff[level] -= len(b)
	a.levelCount[level] += len(b)
	a.length += len(b)
	copy(a.bytes[absindex-len(b):], b)
}

func (a *Array) rebalance() {
	for l := 0; l < a.levels(); l++ {
		if a.needsBalance(l) {
			// Time to spill downward
			if l == a.levels()-1 {
				a.createNewLevel()
			}
			overlow := a.levelCount[l] - (a.low << l)
			if overlow < 0 {
				panic(fmt.Sprintf("l: %d, count %#v, off %#v", l, a.levelCount, a.levelOff))
			}
			toMove := min(overlow, a.levelFree(l+1))
			a.levelOff[l+1] -= toMove
			copy(a.bytes[a.levelOff[l+1]:], a.bytes[a.levelStart(l+1)-toMove:a.levelStart(l+1)])
			a.levelCount[l+1] += toMove
			copy(a.bytes[a.levelOff[l]+toMove:a.levelStart(l+1)], a.bytes[a.levelOff[l]:])
			a.levelOff[l] += toMove
			a.levelCount[l] -= toMove
		}
	}
}

func (a *Array) createNewLevel() {
	newLevel := a.levels()
	a.bytes = append(a.bytes, make([]byte, a.levelTotalCapacity(newLevel))...)
	a.levelCount = append(a.levelCount, 0)
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

func (a *Array) needsBalance(l int) bool {
	return a.levelCount[l] > (a.highwater << l)
}

func (a *Array) Len() int {
	return a.length
}

// findOffset finds the real offset in Array.bytes
// that corresponds with the abstracted offset
// as well as the level at which that offset occurs
func (a *Array) findOffset(idx int) (level, offset int) {
	var i, x int
	for i, x = range a.levelCount {
		if idx == 0 {
			return i, a.levelOff[i]
		}
		if x > idx {
			return i, a.levelOff[i] + idx
		}
		idx -= x
	}
	if idx == 0 {
		return i, a.levelOff[i] + a.levelCount[i]
	}
	panic(fmt.Sprintf("offset too large %d", idx))
}

func (a *Array) levelTotalCapacity(l int) int {
	return a.pagesize << l
}
func (a *Array) levelStart(l int) int {
	return (a.pagesize << l) - a.pagesize
}

func (a *Array) levelUsage(l int) int {
	return a.levelCount[l]
}

func (a *Array) levelFree(l int) int {
	return a.levelOff[l] - a.levelStart(l)
}

func (a *Array) levels() int {
	return len(a.levelCount)
}

func (a *Array) checkInvariants() {
	s := 0
	for _, x := range a.levelCount {
		s += x
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
