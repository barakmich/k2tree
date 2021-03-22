package k2tree

import (
	"fmt"
	"math/bits"

	"github.com/barakmich/k2tree/bytearray"
)

type byteArray struct {
	bytes  bytearray.ByteArray
	length int
	total  int
}

var _ bitarray = (*byteArray)(nil)

func newByteArray(bytes bytearray.ByteArray) *byteArray {
	return &byteArray{
		bytes:  bytes,
		length: 0,
		total:  0,
	}
}

func (b *byteArray) Len() int {
	return b.length
}

func (b *byteArray) Set(at int, val bool) {
	if at >= b.length {
		panic("can't set a bit beyond the size of the array")
	}
	off := at >> 3
	bit := byte(at & 0x07)
	t := byte(0x01 << (7 - bit))
	orig := b.bytes.Get(off)
	var newbyte byte
	if val {
		newbyte = orig | t
	} else {
		newbyte = orig &^ t
	}
	b.bytes.Set(off, newbyte)
	if newbyte != orig {
		if val {
			b.total++
		} else {
			b.total--
		}
	}
}

func (b *byteArray) Count(from, to int) int {
	if from > to {
		from, to = to, from
	}
	if from > b.length || to > b.length {
		panic("out of range")
	}
	if from == to {
		return 0
	}
	c := 0
	startoff := from >> 3
	startbit := byte(from & 0x07)
	endoff := to >> 3
	endbit := byte(to & 0x07)
	if startoff == endoff {
		abit := byte(0xFF >> startbit)
		bbit := byte(0xFF >> endbit)
		return bits.OnesCount8(b.bytes.Get(startoff) & (abit &^ bbit))
	}
	if startbit != 0 {
		c += bits.OnesCount8(b.bytes.Get(startoff) & (0xFF >> startbit))
		startoff++
	}
	if endbit != 0 {
		c += bits.OnesCount8(b.bytes.Get(endoff) & (0xFF &^ (0xFF >> endbit)))
	}
	c += int(b.bytes.PopCount(startoff, endoff))
	return c
}

func (b *byteArray) Total() int {
	return b.total
}

func (b *byteArray) Get(at int) bool {
	off := at >> 3
	lowb := byte(at & 0x07)
	mask := byte(0x01 << (7 - lowb))
	return !(b.bytes.Get(off)&mask == 0x00)
}

func (b *byteArray) String() string {
	str := fmt.Sprintf("L%d T%d ", b.length, b.total)
	return str
}

func (b *byteArray) debug() string {
	return b.String()
}

func (b *byteArray) Insert(n, at int) (err error) {
	if at > b.length {
		panic("can't extend starting at a too large offset")
	}
	if n == 0 {
		return nil
	}
	if at%4 != 0 {
		panic("can only insert a sliceArray at offset multiples of 4")
	}
	if n%8 == 0 {
		err = b.insertEight(n, at)
	} else if n == 4 {
		err = b.insertFour(at)

	} else if n%4 == 0 {
		mult8 := (n >> 3) << 3
		err = b.insertEight(mult8, at)
		if err != nil {
			return err
		}
		err = b.insertFour(at)
	} else {
		panic("can only extend a sliceArray by nibbles or multiples of 8")
	}
	if err != nil {
		return err
	}
	b.length = b.length + n
	return nil
}

func (b *byteArray) insertFour(at int) error {
	if b.length%8 == 0 {
		// We need more space
		b.bytes.Insert(b.bytes.Len(), []byte{0x00})
	}
	off := at >> 3
	var inbyte byte
	if at%8 != 0 {
		inbyte = b.bytes.Get(off)
		b.bytes.Set(off, inbyte&0xF0)
		off++
	}
	inbyte = inbyte << 4
	for i := off; i < b.bytes.Len(); i++ {
		t := b.bytes.Get(i)
		b.bytes.Set(i, t>>4|inbyte)
		inbyte = t << 4
	}
	if inbyte != 0x00 {
		panic("Overshot")
	}
	return nil
}

func (b *byteArray) insertEight(n, at int) error {

	nBytes := n >> 3
	newbytes := make([]byte, nBytes)
	if at == b.length {
		b.bytes.Insert(b.bytes.Len(), newbytes)
		return nil
	}

	off := at >> 3
	if at%8 == 0 {
		b.bytes.Insert(off, newbytes)
	} else {
		b.bytes.Insert(off+1, newbytes)
		oldoff := b.bytes.Get(off)
		b.bytes.Set(off+nBytes, oldoff&0x0F)
		b.bytes.Set(off, oldoff&0xF0)
	}
	return nil
}
