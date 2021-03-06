package bytearray

import "github.com/tmthrgd/go-popcount"

var _ ByteArray = &FrontSlice{}

type FrontSlice struct {
	bytes []byte
	off   int
}

func NewFrontSlice(pagesize int) *FrontSlice {
	return &FrontSlice{
		bytes: make([]byte, pagesize),
		off:   pagesize,
	}
}

func (f *FrontSlice) Len() int {
	return len(f.bytes) - f.off
}

func (f *FrontSlice) Set(idx int, b byte) {
	f.bytes[idx+f.off] = b
}

func (f *FrontSlice) Get(idx int) byte {
	return f.bytes[idx+f.off]
}

func (f *FrontSlice) Insert(idx int, b []byte) {
	if f.off-len(b) < 0 {
		oldlen := len(f.bytes)
		newbytes := make([]byte, oldlen*2)
		copy(newbytes[oldlen:], f.bytes)
		f.bytes = newbytes
		f.off = oldlen
		f.Insert(idx, b)
		return
	}
	copy(f.bytes[f.off-len(b):], f.bytes[f.off:f.off+idx])
	f.off -= len(b)
	copy(f.bytes[f.off+idx:], b)
}

func (f *FrontSlice) PopCount(start, end int) uint64 {
	return popcount.CountBytes(f.bytes[f.off+start : f.off+end])
}

func (f *FrontSlice) Copy(from, to, n int) {
	copy(f.bytes[f.off+to:], f.bytes[f.off+from:f.off+from+n])
}
