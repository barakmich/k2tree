package spillover

var _ testArray = &frontSliceTest{}

type frontSliceTest struct {
	bytes []byte
	off   int
}

func newFrontSliceTest(pagesize int) *frontSliceTest {
	return &frontSliceTest{
		bytes: make([]byte, pagesize),
		off:   pagesize,
	}
}

func (f *frontSliceTest) Len() int {
	return len(f.bytes) - f.off
}

func (f *frontSliceTest) Set(idx int, b byte) {
	f.bytes[idx+f.off] = b
}

func (f *frontSliceTest) Get(idx int) byte {
	return f.bytes[idx+f.off]
}

func (f *frontSliceTest) Insert(idx int, b []byte) {
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
