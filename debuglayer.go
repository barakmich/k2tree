package k2tree

import "fmt"

type debugArray struct {
	bitarray
}

var _ bitarray = (*debugArray)(nil)

func newDebugArray(bits bitarray) *debugArray {
	return &debugArray{
		bitarray: bits,
	}
}

func (d *debugArray) Count(from, to int) int {
	fmt.Printf("** Count from: %d to: %d ", from, to)
	res := d.bitarray.Count(from, to)
	fmt.Printf("return: %d\n", res)
	return res
}

func (d *debugArray) Insert(n, at int) error {
	fmt.Printf("** Insert n: %d at: %d\n", n, at)
	return d.bitarray.Insert(n, at)
}

func (d *debugArray) Set(n int, val bool) {
	fmt.Printf("** Set n: %d val: %v\n", n, val)
	d.bitarray.Set(n, val)
}

func (d *debugArray) debug() string {
	return fmt.Sprintf("DebugArray\n%s", d.bitarray.debug())
}
