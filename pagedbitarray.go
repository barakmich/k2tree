package k2tree

import (
	"fmt"
	"math"
	"math/bits"

	popcount "github.com/tmthrgd/go-popcount"
)

type pagedBitarray struct {
	pages         [][]byte
	firstLevelLen int
	levelLength   []int
	levelTree     []int
	bytelength    int
	bitlength     int
	pagesize      int
	high          int
	low           int
	bittotal      int
}

var _ bitarray = (*pagedBitarray)(nil)

func (p *pagedBitarray) Len() int {
	return p.bitlength
}

func (p *pagedBitarray) Total() int {
	return p.bittotal
}

func (p *pagedBitarray) debug() string {
	str := fmt.Sprintf("L%d T%d ", p.bitlength, p.bittotal)
	return str
}

func (p *pagedBitarray) Set(at int, val bool) {
	if at >= p.bitlength {
		panic("can't set a bit beyond the size of the array")
	}
	off := at >> 3
	bit := byte(at & 0x07)
	t := byte(0x01 << (7 - bit))
	level, byteoff := p.findOffset(off)
	orig := p.pages[level][byteoff]
	var newbyte byte
	if val {
		newbyte = orig | t
	} else {
		newbyte = orig &^ t
	}
	p.pages[level][byteoff] = newbyte
	if newbyte != orig {
		if val {
			p.bittotal++
		} else {
			p.bittotal--
		}
	}
}

func (p *pagedBitarray) Count(from, to int) int {
	if from > to {
		from, to = to, from
	}
	if from == to {
		return 0
	}
	var c uint64
	start := from >> 3
	startbit := byte(from & 0x07)
	end := to >> 3
	endbit := byte(to & 0x07)

	startl, startoff := p.findOffset(start)

	if start == end {
		abit := byte(0xFF >> startbit)
		bbit := byte(0xFF >> endbit)
		return bits.OnesCount8(p.pages[startl][startoff] & (abit &^ bbit))
	}

	delta := end - start
	if startbit != 0 {
		c += uint64(bits.OnesCount8(p.pages[startl][startoff] & (0xFF >> startbit)))
		startoff++
		delta--
		if startoff == p.levelLength[startl] {
			startl += 1
			startoff = 0
			if startl == len(p.pages) {
				return int(c)
			}
		}
	}

	var endl, endoff int
	if startoff+delta < p.levelLength[startl] {
		endl, endoff = startl, (startoff + delta)
	} else {
		endl, endoff = p.findOffset(end)
	}

	if endbit != 0 {
		c += uint64(bits.OnesCount8(p.pages[endl][endoff] & (0xFF &^ (0xFF >> endbit))))
	}

	if startl == endl {
		c += popcount.CountBytes(p.pages[startl][startoff:endoff])
		return int(c)
	}

	c += popcount.CountBytes(p.pages[startl][startoff:p.levelLength[startl]])
	c += popcount.CountBytes(p.pages[endl][:endoff])
	for l := startl + 1; l < endl; l++ {
		c += popcount.CountBytes(p.pages[l][:p.levelLength[l]])
	}
	return int(c)
}

func (p *pagedBitarray) Get(at int) bool {
	off := at >> 3
	lowb := byte(at & 0x07)
	mask := byte(0x01 << (7 - lowb))
	level, idx := p.findOffset(off)
	return !(p.pages[level][idx]&mask == 0x00)
}

func (p *pagedBitarray) Insert(n, at int) (err error) {
	if at > p.bitlength {
		panic("can't extend starting at a too large offset")
	}
	if n == 0 {
		return nil
	}
	if at%4 != 0 {
		panic("can only insert a sliceArray at offset multiples of 4")
	}
	if n%8 == 0 {
		err = p.insertEight(n, at)
	} else if n == 4 {
		err = p.insertFour(at)

	} else if n%4 == 0 {
		mult8 := (n >> 3) << 3
		err = p.insertEight(mult8, at)
		if err != nil {
			return err
		}
		err = p.insertFour(at)
	} else {
		panic("can only extend a sliceArray by nibbles or multiples of 8")
	}
	if err != nil {
		return err
	}
	p.bitlength = p.bitlength + n
	return nil
}

func (p *pagedBitarray) insertFour(at int) error {
	if p.bitlength%8 == 0 {
		// We need more space
		p.insertBytes(p.bytelength, []byte{0x00})
	}
	off := at >> 3
	var inbyte byte

	level, byteoff := p.findOffset(off)
	if at%8 != 0 {
		inbyte = p.pages[level][byteoff]
		p.pages[level][byteoff] = inbyte & 0xF0
		byteoff++
		if byteoff == len(p.pages[level]) {
			level += 1
			byteoff = 0
		}
	}

	inbyte = inbyte << 4

	for l := level; l < len(p.pages); l++ {
		inbyte = insertFourBits(p.pages[l][byteoff:p.levelLength[l]], inbyte)
		byteoff = 0
	}
	if inbyte != 0x00 {
		panic("Overshot")
	}
	return nil
}

func (p *pagedBitarray) insertEight(n, at int) error {
	nBytes := n >> 3
	newbytes := make([]byte, nBytes)
	if at == p.bitlength {
		p.insertBytes(p.bytelength, newbytes)
		return nil
	}

	off := at >> 3
	if at%8 == 0 {
		p.insertBytes(off, newbytes)
	} else {
		p.insertBytes(off+1, newbytes)
		oldoff := p.getByte(off)
		p.setByte(off+nBytes, oldoff&0x0F)
		p.setByte(off, oldoff&0xF0)
	}
	return nil
}

func newPagedBitarray(pagesize int, highwaterPercentage, lowUtilization float64) *pagedBitarray {
	if highwaterPercentage < lowUtilization {
		panic("User error: highwaterPercentage is higher than lowUtilization")
	}
	hw := int(math.Round(highwaterPercentage * float64(pagesize)))
	low := int(math.Round(lowUtilization * float64(pagesize)))
	pages := make([][]byte, 1)
	pages[0] = make([]byte, pagesize)

	return &pagedBitarray{
		pages:       pages,
		bytelength:  0,
		bitlength:   0,
		levelLength: []int{0},
		// Linear tree in [root, left, right] order, recursively.
		// So, a perfect binary tree of {1-7} with 4 on the top would be
		// 4, 2, 6, 1, 3, 5, 7
		// Keeps the counts to how many set bytes are used before that level
		// in the array.
		levelTree: []int{0},
		pagesize:  pagesize,
		high:      hw,
		low:       low,
		bittotal:  0,
	}
}

func (p *pagedBitarray) updateTree(level, delta int) {
	var req int // Bits required to represent all the levels
	if p.levels() == 1 {
		req = 1
	} else {
		req = bits.Len64(uint64(p.levels() - 1))
	}
	treeidx := 0
	for req > 0 {
		isEmpty := (level & (0x1 << (req - 1))) == 0
		if isEmpty {
			p.levelTree[treeidx] += delta
			treeidx = (treeidx << 1) + 1
		} else {
			treeidx = (treeidx << 1) + 2
		}
		req--
	}
}

func (p *pagedBitarray) setByte(idx int, b byte) {
	level, off := p.findOffset(idx)
	p.pages[level][off] = b
}

func (p *pagedBitarray) getByte(idx int) byte {
	level, off := p.findOffset(idx)
	return p.pages[level][off]
}

func (p *pagedBitarray) findOffset(idx int) (level int, offset int) {
	if idx == p.bytelength {
		return p.levels() - 1, p.levelLength[len(p.pages)-1]
	}
	if idx < p.firstLevelLen {
		return 0, idx
	}
	tree := p.levelTree
	t := 0
	level = 0
	max := len(tree)
	for t < max {
		level = level << 1
		val := tree[t]
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

func (p *pagedBitarray) insertBytes(idx int, b []byte) {
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

func (p *pagedBitarray) insertIntoLevel(level int, idx int, b []byte) {
	amt := len(b)
	copy(p.pages[level][idx+amt:p.levelLength[level]+amt], p.pages[level][idx:p.levelLength[level]])
	copy(p.pages[level][idx:], b)
	p.bytelength += amt
	p.levelLength[level] += amt
	if level == 0 {
		p.firstLevelLen += amt
	}
	p.updateTree(level, len(b))
}

func (p *pagedBitarray) levels() int {
	return len(p.pages)
}

func (p *pagedBitarray) rebalance() {
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
			if l == 0 {
				p.firstLevelLen -= toMove
			}
			p.updateTree(l, -toMove)
		}
	}
}

func (p *pagedBitarray) createNewLevel() {
	newLevel := p.levels()
	if newLevel != 1 {
		h := bits.Len64(uint64(newLevel))
		if h > bits.Len64(uint64(newLevel-1)) {
			newCum := make([]int, (len(p.levelTree)*2)+1)
			i := 1
			off := 1
			for len(p.levelTree) > 0 {
				copy(newCum[off:], p.levelTree[0:i])
				p.levelTree = p.levelTree[i:]
				i = i << 1
				off += i
			}
			copy(newCum[1:], p.levelTree)
			p.levelTree = newCum
			p.levelTree[0] = p.bytelength
		}
	}
	p.pages = append(p.pages, make([]byte, p.pagesize))
	p.levelLength = append(p.levelLength, 0)
}

func (p *pagedBitarray) levelFree(l int) int {
	return p.pagesize - p.levelLength[l]
}

func (p *pagedBitarray) needsBalance(l int) bool {
	return p.levelLength[l] > p.high
}
