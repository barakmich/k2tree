package k2tree

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"
	"os"

	mmap "github.com/barakmich/mmap-go"
	popcount "github.com/tmthrgd/go-popcount"
)

type mmapArray struct {
	bytes   mmap.MMap
	length  int
	total   int
	file    *os.File
	filelen int
}

const headerBytes = 256

var magicHeader = []byte{'K', '2', 'B', 'A', '1'}

var _ bitarray = (*mmapArray)(nil)

func (s *mmapArray) Len() int {
	return s.length
}

type header struct {
	Magic  [5]byte
	Length int64
	Total  int64
}

func createMMapArray(filename string) (*mmapArray, error) {
	fmt.Println("Createing")
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	err = f.Truncate(headerBytes)
	if err != nil {
		return nil, err
	}
	_, err = f.WriteAt(magicHeader, 0)
	if err != nil {
		return nil, err
	}
	err = f.Close()
	if err != nil {
		return nil, err
	}
	return openMMapArray(filename)
}

func (s *mmapArray) Close() error {
	var h header
	for i, b := range magicHeader {
		h.Magic[i] = b
	}
	h.Length = int64(s.length)
	h.Total = int64(s.total)
	s.file.Seek(0, 0)
	binary.Write(s.file, binary.BigEndian, &h)
	err := s.bytes.Unmap()
	if err != nil {
		return err
	}
	return s.file.Close()
}

func openMMapArray(filename string) (*mmapArray, error) {
	f, err := os.OpenFile(filename, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	var h header
	err = binary.Read(f, binary.BigEndian, &h)
	if err != nil {
		return nil, err
	}
	for i, b := range magicHeader {
		if b != h.Magic[i] {
			return nil, errors.New("incompatible magic header")
		}
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	f.Seek(0, 0)
	m, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		return nil, err
	}
	return &mmapArray{
		bytes:   m,
		length:  int(h.Length),
		total:   int(h.Total),
		file:    f,
		filelen: int(fi.Size()),
	}, nil
}

func newMMapArray(filename string) (*mmapArray, error) {
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return createMMapArray(filename)
		}
		return nil, err
	}
	return openMMapArray(filename)
}

func (s *mmapArray) Set(at int, val bool) {
	if at >= s.length {
		panic("can't set a bit beyond the size of the array")
	}
	off := at>>3 + headerBytes
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

func (s *mmapArray) Count(from, to int) int {
	if from > to {
		from, to = to, from
	}
	if from > s.length || to > s.length {
		panic("out of range")
	}
	c := 0
	startoff := from>>3 + headerBytes
	startbit := byte(from & 0x07)
	endoff := to>>3 + headerBytes
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

func (s *mmapArray) Total() int {
	return s.total
}

func (s *mmapArray) Get(at int) bool {
	off := at>>3 + headerBytes
	b := byte(at & 0x07)
	mask := byte(0x01 << (7 - b))
	return !(s.bytes[off]&mask == 0x00)
}

func (s *mmapArray) String() string {
	str := fmt.Sprintf("%d ", s.length)
	for _, x := range s.bytes[headerBytes:] {
		str += fmt.Sprintf("%08b ", x)
	}
	return str
}

func (s *mmapArray) debug() string {
	return s.String()
}

func (s *mmapArray) Insert(n, at int) (err error) {
	if at > s.length {
		panic("can't extend starting at a too large offset")
	}
	if n == 0 {
		return nil
	}
	if n%4 != 0 {
		panic("can only extend a mmapArray by nibbles (multiples of 4)")
	}
	if at%4 != 0 {
		panic("can only insert a mmapArray at offset multiples of 4")
	}
	if n%8 == 4 {
		err = s.insertFour(n, at)
	} else {
		err = s.insertEight(n, at)
	}
	s.length = s.length + n
	return err
}

func (s *mmapArray) insertFour(n, at int) error {
	if s.length%8 == 4 {
		// We have some extra bits
		return s.insertFourExtra(n, at)
	}
	newbytes := (n >> 3) + 1
	err := s.extendBytes(newbytes)
	if err != nil {
		return err
	}
	if at == s.length {
		return nil
	}

	off := at>>3 + headerBytes
	if at%8 == 4 {
		copy(s.bytes[off+newbytes:], s.bytes[off+1:])
		for x := 1; x < newbytes; x++ {
			s.bytes[off+x] = 0x00
		}
		a := s.bytes[off] << 4
		s.bytes[off] &= 0xF0
		for x := off + newbytes; x < len(s.bytes); x++ {
			t := s.bytes[x] << 4
			b := s.bytes[x] >> 4
			s.bytes[x] = a | b
			a = t
		}
	} else {
		if newbytes != 0 {
			copy(s.bytes[off+newbytes:], s.bytes[off:])
			for x := 0; x < newbytes; x++ {
				s.bytes[off+x] = 0x00
			}
		}
		for x := off + newbytes - 1; x < len(s.bytes)-1; x++ {
			b := s.bytes[x+1] >> 4
			s.bytes[x] = s.bytes[x]<<4 | b
		}
		s.bytes[len(s.bytes)-1] = s.bytes[len(s.bytes)-1] << 4
	}
	return nil
}

func (s *mmapArray) insertFourExtra(n, at int) error {
	newbytes := (n - 4) >> 3
	err := s.extendBytes(newbytes)
	if err != nil {
		return err
	}
	if at == s.length {
		return nil
	}
	off := at>>3 + headerBytes
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

func (s *mmapArray) insertEight(n, at int) error {

	nBytes := n >> 3
	err := s.extendBytes(nBytes)
	if err != nil {
		return err
	}
	if at == s.length {
		return nil
	}

	off := at>>3 + headerBytes
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

func (s *mmapArray) extendBytes(n int) error {
	s.filelen += n
	return s.bytes.Truncate(s.file, s.filelen)
}
