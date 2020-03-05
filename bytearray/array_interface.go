package bytearray

type testArray interface {
	Len() int
	Set(idx int, b byte)
	Get(idx int) byte
	Insert(idx int, b []byte)
}

var _ testArray = &Array{}
var _ testArray = &sliceTest{}

type sliceTest struct {
	bytes []byte
}

func newSliceTest() *sliceTest {
	return &sliceTest{}
}

func (s *sliceTest) Len() int {
	return len(s.bytes)
}

func (s *sliceTest) Set(idx int, b byte) {
	s.bytes[idx] = b
}

func (s *sliceTest) Get(idx int) byte {
	return s.bytes[idx]
}

func (s *sliceTest) Insert(idx int, b []byte) {
	newbytes := len(b)
	s.bytes = append(s.bytes, make([]byte, newbytes)...)
	copy(s.bytes[idx+newbytes:], s.bytes[idx:])
	for x := 0; x < newbytes; x++ {
		s.bytes[idx+x] = b[x]
	}
}
