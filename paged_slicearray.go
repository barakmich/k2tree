package k2tree

import "fmt"

type pagedSliceArray struct {
	arrays   []*sliceArray
	pagesize int
}

var _ bitarray = (*pagedSliceArray)(nil)

var maxPageSize = 1000

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
		n += x.Len()
	}
	return n
}

func (p *pagedSliceArray) Set(at int, val bool) {
	for _, x := range p.arrays {
		if x.Len() > at {
			x.Set(at, val)
			return
		}
		at -= x.Len()
	}
}

func (p *pagedSliceArray) Get(at int) bool {
	for _, x := range p.arrays {
		if x.Len() > at {
			return x.Get(at)
		}
		at -= x.Len()
	}
	panic("end of arrays")
}

func (p *pagedSliceArray) Count(from, to int) int {
	n := to - from
	count := 0
	for _, x := range p.arrays {
		if from >= x.Len() {
			from -= x.Len()
			continue
		}
		if n <= (x.Len() - from) {
			count += x.Count(from, from+n)
			return count
		} else {
			n -= x.Len() - from
			count += x.Count(from, x.Len())
			from = 0
		}
	}
	panic("end of arrays")
}

func (p *pagedSliceArray) Insert(n int, at int) error {
	var page *sliceArray
	var pagei int
	origat := at
	for i, x := range p.arrays {
		if x.Len() >= at {
			pagei = i
			page = x
			break
		}
		at -= x.Len()
	}
	if page.Len() < p.pagesize {
		return page.Insert(n, at)
	}
	l := len(page.bytes) / 2
	newpage := &sliceArray{
		bytes:  page.bytes[:l],
		length: l * 8,
	}
	page.bytes = page.bytes[l:]
	page.length -= l * 8
	p.arrays = append(p.arrays[:pagei], append([]*sliceArray{newpage}, p.arrays[pagei:]...)...)
	return p.Insert(n, origat)
}

func (p *pagedSliceArray) debug() string {
	s := ""
	for i, x := range p.arrays {
		s += fmt.Sprintf("Array[%d]:\n%s\n", i, x.debug())
	}
	return s
}
