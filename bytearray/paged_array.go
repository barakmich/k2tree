package bytearray

import (
	"bytes"
	"fmt"
	"math"
	"math/bits"

	"github.com/tmthrgd/go-popcount"
)

type PagedArray struct {
	pages       [][]byte
	levelLength []int
	levelCum    []int
	length      int
	pagesize    int
	high        int
	low         int
}

func NewPaged(pagesize int, highwaterPercentage, lowUtilization float64) *PagedArray {
	if highwaterPercentage < lowUtilization {
		panic("User error: highwaterPercentage is higher than lowUtilization")
	}
	hw := int(math.Round(highwaterPercentage * float64(pagesize)))
	low := int(math.Round(lowUtilization * float64(pagesize)))
	pages := make([][]byte, 1)
	pages[0] = make([]byte, pagesize)

	return &PagedArray{
		pages:       pages,
		length:      0,
		levelLength: []int{0},
		levelCum:    []int{0},
		pagesize:    pagesize,
		high:        hw,
		low:         low,
	}
}

func (p *PagedArray) updateTree(level, delta int) {
	var req int
	if p.levels() == 1 {
		req = 1
	} else {
		req = bits.Len64(uint64(p.levels() - 1))
	}
	treeidx := 0
	for req > 0 {
		isEmpty := (level & (0x1 << (req - 1))) == 0
		if isEmpty {
			p.levelCum[treeidx] += delta
			treeidx = (treeidx << 1) + 1
		} else {
			treeidx = (treeidx << 1) + 2
		}
		req--
	}
}

func (p *PagedArray) Len() int {
	return p.length
}

func (p *PagedArray) Set(idx int, b byte) {
	if idx >= p.length {
		panic("Writing off the edge of the Paged Array")
	}
	level, off := p.findOffset(idx)
	p.pages[level][off] = b
}

func (p *PagedArray) Get(idx int) byte {
	if idx >= p.length {
		panic("Fetching off the edge of the Paged Array")
	}
	level, off := p.findOffset(idx)
	return p.pages[level][off]
}

func (p *PagedArray) findOffset(idx int) (level int, offset int) {
	if idx > p.length {
		panic("offset too large")
	}
	if idx == p.length {
		return p.levels() - 1, p.levelLength[len(p.pages)-1]
	}
	if idx == 0 {
		return 0, 0
	}
	t := 0
	level = 0
	for t < len(p.levelCum) {
		level = level << 1
		val := p.levelCum[t]
		if idx >= val {
			idx -= val
			level |= 0x1
			t = (t << 1) + 2
		} else {
			t = (t << 1) + 1
		}
	}
	return level, idx
}

func (p *PagedArray) Insert(idx int, b []byte) {
	for len(b) != 0 {
		l, off := p.findOffset(idx)
		var toInsert []byte
		free := p.levelFree(l)
		if len(b) <= free {
			toInsert = b
			b = nil
		} else {
			toInsert = b[:free]
			b = b[free:]
		}
		p.insertIntoLevel(l, off, toInsert)
		if p.needsBalance(l) {
			p.rebalance()
		}
	}
}

func (p *PagedArray) insertIntoLevel(level int, idx int, b []byte) {
	amt := len(b)
	copy(p.pages[level][idx+amt:p.levelLength[level]+amt], p.pages[level][idx:p.levelLength[level]])
	copy(p.pages[level][idx:], b)
	p.length += amt
	p.levelLength[level] += amt
	p.updateTree(level, len(b))
}

func (p *PagedArray) levels() int {
	return len(p.pages)
}

func (p *PagedArray) rebalance() {
	for l := 0; l < p.levels(); l++ {
		if p.needsBalance(l) {
			// Time to spill downward
			if l == p.levels()-1 {
				p.createNewLevel()
			}
			overlow := p.levelLength[l] - p.low
			if overlow < 0 {
				panic(fmt.Sprintf("l: %d, is under low water", l))
			}
			toMove := min(overlow, p.levelFree(l+1))
			copy(p.pages[l+1][toMove:], p.pages[l+1][:p.levelLength[l+1]])
			copy(p.pages[l+1][:toMove], p.pages[l][p.levelLength[l]-toMove:p.levelLength[l]])
			p.levelLength[l+1] += toMove
			p.updateTree(l+1, toMove)
			p.levelLength[l] -= toMove
			p.updateTree(l, -toMove)
		}
	}
}

func (p *PagedArray) createNewLevel() {
	newLevel := p.levels()
	if newLevel != 1 {
		h := bits.Len64(uint64(newLevel))
		if h > bits.Len64(uint64(newLevel-1)) {
			newCum := make([]int, (len(p.levelCum)*2)+1)
			i := 1
			off := 1
			for len(p.levelCum) > 0 {
				copy(newCum[off:], p.levelCum[0:i])
				p.levelCum = p.levelCum[i:]
				i = i << 1
				off += i
			}
			copy(newCum[1:], p.levelCum)
			p.levelCum = newCum
			p.levelCum[0] = p.length
		}
	}
	p.pages = append(p.pages, make([]byte, p.pagesize))
	p.levelLength = append(p.levelLength, 0)
}

func (p *PagedArray) levelFree(l int) int {
	return p.pagesize - p.levelLength[l]
}

func (p *PagedArray) needsBalance(l int) bool {
	return p.levelLength[l] > p.high
}

func (p *PagedArray) PopCount(start int, end int) uint64 {
	var count uint64
	startl, startoff := p.findOffset(start)
	endl, endoff := p.findOffset(end)

	if startl == endl {
		return popcount.CountBytes(p.pages[startl][startoff:endoff])
	}

	count += popcount.CountBytes(p.pages[startl][startoff:p.levelLength[startl]])
	count += popcount.CountBytes(p.pages[endl][:endoff])
	for l := startl + 1; l < endl; l++ {
		count += popcount.CountBytes(p.pages[l][:p.levelLength[l]])
	}
	return count
}

func (p *PagedArray) Copy(from int, to int, n int) {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Grow(n)

	// Copy into the buffer
	startp, startoff := p.findOffset(from)
	endp, endoff := p.findOffset(from + n)

	if startp == endp {
		buf.Write(p.pages[startp][startoff:endoff])
	} else {
		buf.Write(p.pages[startp][startoff:p.levelLength[startp]])
		for l := startp + 1; l < endp; l++ {
			buf.Write(p.pages[l][:p.levelLength[l]])
		}
		buf.Write(p.pages[endp][:endoff])
	}

	// Copy out of the buffer
	startp, startoff = p.findOffset(to)
	endp, endoff = p.findOffset(to + n)
	if startp == endp {
		copied, err := buf.Read(p.pages[startp][startoff:endoff])
		if err != nil {
			panic(err)
		}
		if n != copied {
			panic("didn't copy everything?")
		}
	} else {
		copied := 0
		read, err := buf.Read(p.pages[startp][startoff:p.levelLength[startp]])
		if err != nil {
			panic(err)
		}
		copied += read
		for l := startp + 1; l < endp; l++ {
			read, err := buf.Read(p.pages[l][:p.levelLength[l]])
			if err != nil {
				panic(err)
			}
			copied += read
		}
		read, err = buf.Read(p.pages[endp][:endoff])
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
