package bytearray

import (
	"bytes"
	"sync"
)

var (
	bufPool = &sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
)
