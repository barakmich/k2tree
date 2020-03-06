package k2tree

import (
	"fmt"
	"math/bits"

	"git.barakmich.com/barak/k2tree/bytearray"
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
	if n%4 != 0 {
		panic("can only extend a sliceArray by nibbles (multiples of 4)")
	}
	if at%4 != 0 {
		panic("can only insert a sliceArray at offset multiples of 4")
	}
	if n%8 == 4 {
		err = b.insertFour(n, at)
	} else {
		err = b.insertEight(n, at)
	}
	b.length = b.length + n
	return err
}

func (b *byteArray) insertFour(n, at int) error {
	if b.length%8 == 4 {
		// We have some extra bits
		return b.insertFourExtra(n, at)
	}
	newbytesN := (n >> 3) + 1
	newbytes := make([]byte, newbytesN)

	if at == b.length {
		b.bytes.Insert(b.bytes.Len(), newbytes)
		return nil
	}

	off := at >> 3
	if at%8 == 4 {
		b.bytes.Insert(off+1, newbytes)
		byteAtOff := b.bytes.Get(off)
		a := byteAtOff << 4
		b.bytes.Set(off, byteAtOff&0xF0)
		for x := off + 1 + newbytesN; x < b.bytes.Len(); x++ {
			byteAtOff = b.bytes.Get(x)
			t := byteAtOff << 4
			u := byteAtOff >> 4
			b.bytes.Set(x-1, a|u)
			a = t
		}
	} else {
		if newbytesN != 0 {
			b.bytes.Insert(off, newbytes)
		}
		for x := off + newbytesN - 1; x < b.bytes.Len()-1; x++ {
			u := b.bytes.Get(x+1) >> 4
			b.bytes.Set(x, b.bytes.Get(x)<<4|u)
		}
		b.bytes.Set(b.bytes.Len()-1, b.bytes.Get(b.bytes.Len()-1)<<4)
	}
	return nil
}

func (b *byteArray) insertFourExtra(n, at int) error {
	newbytesN := (n - 4) >> 3
	newbytes := make([]byte, newbytesN)
	if at == b.length {
		b.bytes.Insert(b.bytes.Len(), newbytes)
		return nil
	}

	off := at >> 3
	if at%8 == 4 {
		if newbytesN != 0 {
			b.bytes.Insert(off+1, newbytes)
		}
		oldOff := b.bytes.Get(off)
		a := oldOff << 4
		b.bytes.Set(off, oldOff&0xF0)
		for x := off + newbytesN + 1; x < b.bytes.Len(); x++ {
			oldx := b.bytes.Get(x)
			t := oldx << 4
			b.bytes.Set(x, a|(oldx>>4))
			a = t
		}
	} else {
		if newbytesN != 0 {
			b.bytes.Insert(off, newbytes)
		}
		var a byte
		for x := off + newbytesN; x < b.bytes.Len(); x++ {
			oldx := b.bytes.Get(x)
			t := oldx << 4
			b.bytes.Set(x, a|(oldx>>4))
			a = t
		}
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
