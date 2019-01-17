package k2tree

import (
	"fmt"
	"math/bits"
)

type bitarray interface {
	Len() int
	Set(at int, val bool)
	Get(at int) bool
	Count(from, to int) int
	Insert(n int, at int) error
	debug() string
}

var _ bitarray = (*sliceArray)(nil)

type sliceArray struct {
	bytes  []byte
	length int
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
	if val {
		s.bytes[off] = s.bytes[off] | t
	} else {
		s.bytes[off] = s.bytes[off] &^ t
	}
}

func (s *sliceArray) Count(from, to int) int {
	if from > to {
		from, to = to, from
	}
	if from > s.length || to > s.length {
		panic("out of range")
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
	for startoff != endoff {
		c += bits.OnesCount8(s.bytes[startoff])
		startoff++
	}
	return c
}

func (s *sliceArray) Get(at int) bool {
	off := at >> 3
	b := byte(at & 0x07)
	mask := byte(0x01 << (7 - b))
	return !(s.bytes[off]&mask == 0x00)
}

func (s *sliceArray) String() string {
	str := fmt.Sprintf("%d ", s.length)
	for _, x := range s.bytes {
		str += fmt.Sprintf("%08b ", x)
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
	if n%4 != 0 {
		panic("can only extend a sliceArray by nibbles (multiples of 4)")
	}
	if at%4 != 0 {
		panic("can only insert a sliceArray at offset multiples of 4")
	}
	if n%8 == 4 {
		err = s.insertFour(n, at)
	} else {
		err = s.insertEight(n, at)
	}
	s.length = s.length + n
	return err
}

func (s *sliceArray) insertFour(n, at int) error {
	if s.length%8 == 4 {
		// We have some extra bits
		return s.insertFourExtra(n, at)
	}
	newbytes := (n >> 3) + 1
	s.bytes = append(s.bytes, make([]byte, newbytes)...)
	if at == s.length {
		return nil
	}

	off := at >> 3
	if at%8 == 4 {
		copy(s.bytes[off+1+newbytes:], s.bytes[off+1:])
		for x := 0; x < newbytes; x++ {
			s.bytes[off+1+x] = 0x00
		}
		a := s.bytes[off] << 4
		s.bytes[off] &= 0xF0
		for x := off + newbytes; x < len(s.bytes)-1; x++ {
			b := (s.bytes[x+1] & 0xF0) >> 4
			s.bytes[x] = a | b
			a = s.bytes[x+1] << 4
		}
		s.bytes[len(s.bytes)-1] = s.bytes[len(s.bytes)-1] << 4
	} else {
		s.bytes = nil
	}
	return nil
}

func (s *sliceArray) insertFourExtra(n, at int) error {
	newbytes := (n - 4) >> 3
	s.bytes = append(s.bytes, make([]byte, newbytes)...)
	if at == s.length {
		return nil
	}
	off := at >> 3
	if at%8 == 4 {
		if newbytes != 0 {
			copy(s.bytes[off+1+newbytes:], s.bytes[off+1:])
			for x := 0; x < newbytes; x++ {
				s.bytes[off+1+x] = 0x00
			}
		}
		a := s.bytes[off] << 4
		s.bytes[off] &= 0xF0
		for x := off + newbytes + 1; x < len(s.bytes); x++ {
			t := s.bytes[x] << 4
			s.bytes[x] = a | (s.bytes[x] >> 4)
			a = t
		}
	} else {
		if newbytes != 0 {
			copy(s.bytes[off+newbytes:], s.bytes[off:])
			for x := 0; x < newbytes; x++ {
				s.bytes[off+x] = 0x00
			}
		}
		var a byte
		for x := off + newbytes; x < len(s.bytes); x++ {
			t := s.bytes[x] << 4
			s.bytes[x] = a | (s.bytes[x] >> 4)
			a = t
		}
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

//byteExtension := n%8 == 0
//nBytes := n >> 3
//if !byteExtension {
//nBytes++
//}
//s.bytes = append(s.bytes, make([]byte, nBytes)...)

//if at == s.length {
//s.length = s.length + n
//return
//}
//s.length = s.length + n

//off := at >> 3
//if at%8 == 0 {
//copy(s.bytes[off+nBytes:], s.bytes[off:])
//for x := 0; x < nBytes; x++ {
//s.bytes[off+x] = 0x00
//}
//} else {
//copy(s.bytes[off+1+nBytes:], s.bytes[off+1:])
//for x := 0; x < nBytes; x++ {
//s.bytes[off+1+x] = 0x00
//}
//s.bytes[off+nBytes] = s.bytes[off] & 0x0F
//s.bytes[off] = s.bytes[off] & 0xF0
//}
//if byteExtension {
//return
//}
//for x := off + nBytes; x < len(s.bytes)-1; x++ {
//a := s.bytes[x]
//b := s.bytes[x+1]
//s.bytes[x] = a << 4
//s.bytes[x] |= b >> 4
//}
//s.bytes[len(s.bytes)-1] = s.bytes[len(s.bytes)-1] << 4
//}
