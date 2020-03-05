package bytearray

type ByteArray interface {
	Len() int
	Set(idx int, b byte)
	Get(idx int) byte
	Insert(idx int, b []byte)
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
