package bytearray

import "github.com/tmthrgd/go-popcount"

type ByteArray interface {
	Len() int
	Set(idx int, b byte)
	Get(idx int) byte
	Insert(idx int, b []byte)
	PopCount(start, end int) uint64
	Copy(from, to, n int)
}

var _ ByteArray = &SpilloverArray{}
var _ ByteArray = &SliceArray{}

type SliceArray struct {
	bytes []byte
}

func NewSlice() *SliceArray {
	return &SliceArray{}
}

func (s *SliceArray) Len() int {
	return len(s.bytes)
}

func (s *SliceArray) Set(idx int, b byte) {
	s.bytes[idx] = b
}

func (s *SliceArray) Get(idx int) byte {
	return s.bytes[idx]
}

func (s *SliceArray) Insert(idx int, b []byte) {
	newbytes := len(b)
	s.bytes = append(s.bytes, make([]byte, newbytes)...)
	copy(s.bytes[idx+newbytes:], s.bytes[idx:])
	for x := 0; x < newbytes; x++ {
		s.bytes[idx+x] = b[x]
	}
}

func (s *SliceArray) PopCount(start, end int) uint64 {
	return popcount.CountBytes(s.bytes[start:end])
}

func (s *SliceArray) Copy(from, to, n int) {
	copy(s.bytes[to:], s.bytes[from:from+n])
}
