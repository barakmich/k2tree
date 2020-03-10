package bytearray

import (
	"bytes"
	"fmt"
	"math"
	"math/bits"

	"github.com/tmthrgd/go-popcount"
)

type SpilloverArray struct {
	bytes      []byte
	levelOff   []int
	levelCum   []int
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
		levelCum:   []int{0},
		pagesize:   pagesize,
		highwater:  hw,
		low:        low,
		multiplier: multiplier,
	}
}

func (a *SpilloverArray) updateTree(level, delta int) {
	var req int
	if a.levels() == 1 {
		req = 1
	} else {
		req = bits.Len64(uint64(a.levels() - 1))
	}
	treeidx := 0
	for req > 0 {
		isEmpty := (level & (0x1 << (req - 1))) == 0
		if isEmpty {
			a.levelCum[treeidx] += delta
			treeidx = (treeidx << 1) + 1
		} else {
			treeidx = (treeidx << 1) + 2
		}
		req--
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
	a.updateTree(level, len(b))
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
			a.updateTree(l+1, toMove)
			a.updateTree(l, -toMove)
		}
	}
}

func (a *SpilloverArray) createNewLevel() {
	newLevel := a.levels()
	if newLevel != 1 {
		h := bits.Len64(uint64(newLevel))
		if h > bits.Len64(uint64(newLevel-1)) {
			newCum := make([]int, (len(a.levelCum)*2)+1)
			i := 1
			off := 1
			for len(a.levelCum) > 0 {
				copy(newCum[off:], a.levelCum[0:i])
				a.levelCum = a.levelCum[i:]
				i = i << 1
				off += i
			}
			copy(newCum[1:], a.levelCum)
			a.levelCum = newCum
			a.levelCum[0] = a.length
		}
	}
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
	if idx > a.length {
		panic("offset too large")
	}
	if idx == a.length {
		return a.levels() - 1, len(a.bytes)
	}
	if idx == 0 {
		return 0, a.levelOff[0]
	}
	t := 0
	level = 0
	for t < len(a.levelCum) {
		level = level << 1
		val := a.levelCum[t]
		if idx >= val {
			idx -= val
			level |= 0x1
			t = (t << 1) + 2
		} else {
			t = (t << 1) + 1
		}
	}
	return level, a.levelOff[level] + idx
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

func (a *SpilloverArray) Copy(from, to, n int) {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Grow(n)

	// Copy into the buffer
	startl, startoff := a.findOffset(from)
	endl, endoff := a.findOffset(from + n)

	if startl == endl {
		buf.Write(a.bytes[startoff:endoff])
	} else {
		buf.Write(a.bytes[startoff:a.levelStart(startl+1)])
		for l := startl + 1; l < endl; l++ {
			buf.Write(a.bytes[a.levelOff[l]:a.levelStart(l+1)])
		}
		buf.Write(a.bytes[a.levelOff[endl]:endoff])
	}

	// Copy out of the buffer
	startl, startoff = a.findOffset(to)
	endl, endoff = a.findOffset(to + n)
	if startl == endl {
		copied, err := buf.Read(a.bytes[startoff:endoff])
		if err != nil {
			panic(err)
		}
		if n != copied {
			panic("didn't copy everything?")
		}
	} else {
		copied := 0
		read, err := buf.Read(a.bytes[startoff:a.levelStart(startl+1)])
		if err != nil {
			panic(err)
		}
		copied += read
		for l := startl + 1; l < endl; l++ {
			read, err := buf.Read(a.bytes[a.levelOff[l]:a.levelStart(l+1)])
			if err != nil {
				panic(err)
			}
			copied += read
		}
		read, err = buf.Read(a.bytes[a.levelOff[endl]:endoff])
		if err != nil {
			panic(err)
		}
		copied += read
		if n != copied {
			panic("didn't copy everything")
		}
	}

	bufPool.Put(buf)
}
