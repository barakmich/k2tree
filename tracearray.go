package k2tree

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

type traceArray struct {
	bitarray
	data traceData
}

type traceData struct {
	CountCalls           uint64
	CountLengths         uint64
	CountLengthHistogram *twosHistogram
}

var _ bitarray = (*traceArray)(nil)

func newTraceArray(bits bitarray) *traceArray {
	return &traceArray{
		bitarray: bits,
		data: traceData{
			CountLengthHistogram: &twosHistogram{},
		},
	}
}

// Count returns the number of set bits in the interval [from, to).
func (t *traceArray) Count(from int, to int) int {
	t.data.CountCalls += 1
	length := to - from
	t.data.CountLengths += uint64(length)
	//t.data.CountLengthHistogram.Add(length)
	return t.bitarray.Count(from, to)
}

func (t *traceArray) debug() string {
	return fmt.Sprintf("TraceArray\n%s\n%s\n", t.bitarray.debug(), t.data)
}

func (td traceData) String() string {
	return fmt.Sprintf(`
CountCalls: %d
CountLengths: %s
CountLengthHistogram: %s
	`,
		td.CountCalls,
		humanize.Bytes(td.CountLengths/8),
		td.CountLengthHistogram,
	)
}
