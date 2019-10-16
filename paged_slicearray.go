package k2tree

import (
	popcount "github.com/tmthrgd/go-popcount"
)

type pagedSliceArray struct {
	arrays   []*sliceArray
	pagesize int
}

var _ bitarray = (*pagedSliceArray)(nil)

func newPagedSliceArray(size int) *pagedSliceArray {
	arrays := []*sliceArray{
		&sliceArray{},
	}
	return &pagedSliceArray{
		arrays:   arrays,
		pagesize: size,
	}
}

func (p *pagedSliceArray) Len() int {
	n := 0
	for _, x := range p.arrays {
		n += x.length
	}
	return n
}

func (p *pagedSliceArray) Total() int {
	n := 0
	for _, x := range p.arrays {
		n += x.total
	}
	return n
}

func (p *pagedSliceArray) Set(at int, val bool) {
	for _, x := range p.arrays {
		if x.length > at {
			x.Set(at, val)
			return
		}
		at -= x.length
	}
}

func (p *pagedSliceArray) Get(at int) bool {
	for _, x := range p.arrays {
		if x.length > at {
			return x.Get(at)
		}
		at -= x.length
	}
	panic("end of arrays")
}

func (p *pagedSliceArray) Count(from, to int) int {
	n := to - from
	count := 0
	for _, x := range p.arrays {
		length := x.length
		if from >= length {
			from -= length
			continue
		}
		if n <= (length - from) {
			count += x.Count(from, from+n)
			return count
		}

		n -= length - from
		if from != 0 {
			count += x.Count(from, length)
		} else {
			count += x.total
		}
		from = 0
	}
	if n != 0 {
		panic("end of arrays")
	}
	return count
}

func (p *pagedSliceArray) Bytes() []byte {
	var out []byte
	for _, x := range p.arrays {
		out = append(out, x.bytes...)
	}
	return out
}

func (p *pagedSliceArray) Insert(n int, at int) error {
	if at > p.Len() {
		panic("can't extend off the edge of the bitarray")
	}
	var page *sliceArray
	var pagei int
	origat := at
	for i, x := range p.arrays {
		if x.length >= at {
			pagei = i
			page = x
			break
		}
		at -= x.length
	}
	if page.length < p.pagesize {
		return page.Insert(n, at)
	}
	l := len(page.bytes) / 2
	newpage := &sliceArray{
		bytes:  page.bytes[:l],
		length: l * 8,
		total:  int(popcount.CountBytes(page.bytes[:l])),
	}
	page.bytes = page.bytes[l:]
	page.length -= l * 8
	page.total -= newpage.total
	p.arrays = append(p.arrays[:pagei], append([]*sliceArray{newpage}, p.arrays[pagei:]...)...)
	return p.Insert(n, origat)
}

func (p *pagedSliceArray) debug() string {
	s := ""
	//for i, x := range p.arrays {
	//s += fmt.Sprintf("Array[%d]:\n%s\n", i, x.debug())
	//}
	return s
}
