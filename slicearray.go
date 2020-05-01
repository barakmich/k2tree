package k2tree

import (
	"fmt"
	"math/bits"

	popcount "github.com/tmthrgd/go-popcount"
)

type sliceArray struct {
	bytes  []byte
	length int
	total  int
}

var _ bitarray = (*sliceArray)(nil)

func newSliceArray() *sliceArray {
	return &sliceArray{}
}

func (s *sliceArray) Len() int {
	return s.length
}

func (s *sliceArray) Set(at int, val bool) {
	if at >= s.length {
		panic("can't set a bit beyond the size of the array")
	}
	off := at >> 3
	bit := byte(at & 0x07)
	t := byte(0x01 << (7 - bit))
	orig := s.bytes[off]
	if val {
		s.bytes[off] = s.bytes[off] | t
	} else {
		s.bytes[off] = s.bytes[off] &^ t
	}
	if s.bytes[off] != orig {
		if val {
			s.total++
		} else {
			s.total--
		}
	}
}

func (s *sliceArray) Count(from, to int) int {
	if from > to {
		from, to = to, from
	}
	if from > s.length || to > s.length {
		panic("out of range")
	}
	if from == to {
		return 0
	}
	c := 0
	startoff := from >> 3
	startbit := byte(from & 0x07)
	endoff := to >> 3
	endbit := byte(to & 0x07)
	if startoff == endoff {
		a := byte(0xFF >> startbit)
		b := byte(0xFF >> endbit)
		return bits.OnesCount8(s.bytes[startoff] & (a &^ b))
	}
	if startbit != 0 {
		c += bits.OnesCount8(s.bytes[startoff] & (0xFF >> startbit))
		startoff++
	}
	if endbit != 0 {
		c += bits.OnesCount8(s.bytes[endoff] & (0xFF &^ (0xFF >> endbit)))
	}
	c += int(popcount.CountBytes(s.bytes[startoff:endoff]))
	return c
}

func (s *sliceArray) Total() int {
	return s.total
}

func (s *sliceArray) Get(at int) bool {
	off := at >> 3
	b := byte(at & 0x07)
	mask := byte(0x01 << (7 - b))
	return !(s.bytes[off]&mask == 0x00)
}

func (s *sliceArray) String() string {
	str := fmt.Sprintf("L%d T%d ", s.length, s.total)
	for i, x := range s.bytes {
		str += fmt.Sprintf("%08b ", x)
		if i > 20 {
			str += "(first 20)"
			break
		}
	}
	return str
}

func (s *sliceArray) debug() string {
	return s.String()
}

func (s *sliceArray) Insert(n, at int) (err error) {
	if at > s.length {
		panic("can't extend starting at a too large offset")
	}
	if n == 0 {
		return nil
	}
	if at%4 != 0 {
		panic("can only insert a sliceArray at offset multiples of 4")
	}
	if n%8 == 0 {
		err = s.insertEight(n, at)
	} else if n == 4 {
		err = s.insertFour(at)

	} else if n%4 == 0 {
		mult8 := (n >> 3) << 3
		err = s.insertEight(mult8, at)
		if err != nil {
			return err
		}
		err = s.insertFour(at)
	} else {
		panic("can only extend a sliceArray by nibbles or multiples of 8")
	}
	if err != nil {
		return err
	}
	s.length = s.length + n
	return nil
}

func (s *sliceArray) insertFour(at int) error {
	if s.length%8 == 0 {
		// We need more space
		s.bytes = append(s.bytes, 0x00)
	}
	off := at >> 3
	var inbyte byte
	if at%8 != 0 {
		inbyte = s.bytes[off]
		s.bytes[off] &= 0xF0
		off++
	}
	outByte := insertFourBits(s.bytes[off:], inbyte)
	if outByte != 0x00 {
		panic("Overshot")
	}
	return nil
}

func (s *sliceArray) insertEight(n, at int) error {

	nBytes := n >> 3
	s.bytes = append(s.bytes, make([]byte, nBytes)...)
	if at == s.length {
		return nil
	}

	off := at >> 3
	if at%8 == 0 {
		copy(s.bytes[off+nBytes:], s.bytes[off:])
		for x := 0; x < nBytes; x++ {
			s.bytes[off+x] = 0x00
		}
	} else {
		copy(s.bytes[off+1+nBytes:], s.bytes[off+1:])
		for x := 0; x < nBytes; x++ {
			s.bytes[off+1+x] = 0x00
		}
		s.bytes[off+nBytes] = s.bytes[off] & 0x0F
		s.bytes[off] = s.bytes[off] & 0xF0
	}
	return nil
}
