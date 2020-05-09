package k2tree

type Iterator struct {
	tree   *K2Tree
	offset int
	rowcol int
	isRow  bool
}

func newRowIterator(tree *K2Tree, row int) *Iterator {
	return &Iterator{
		tree:   tree,
		offset: -1,
		rowcol: row,
		isRow:  true,
	}
}

func newColumnIterator(tree *K2Tree, col int) *Iterator {
	return &Iterator{
		tree:   tree,
		offset: -1,
		rowcol: col,
		isRow:  false,
	}
}

func (it *Iterator) Next() bool {
	it.offset = it.getNext(it.offset)
	return it.offset != -1
}

func (it *Iterator) Value() int {
	return it.offset
}

func (it *Iterator) getNext(off int) int {
	try := off + 1
	levels := it.tree.levels
	nextval := it.getNextOnLevel(levels, 0, try)
	return nextval
}

func (it *Iterator) getNextOnLevel(level, sublayeroff, val int) int {
	// Invariant: Returned int must be >= val if the value is found or
	// -1 if the function reaches the end of the run of bits.
	if level == 0 {
		return it.getNextOnLeaf(sublayeroff, val)
	}

	startRun := sublayeroff * it.tree.tk.bitsPerLayer
	levelStart := it.tree.levelOffsets[level]
	var offInRun int
	if it.isRow {
		offInRun = it.tree.offsetTForLayer(it.rowcol, val, level)
	} else {
		offInRun = it.tree.offsetTForLayer(val, it.rowcol, level)
	}
	var newoffinrun int

	for {
		bitoff := levelStart + startRun + offInRun
		if it.tree.tbits.Get(bitoff) {
			count := it.tree.tbits.Count(levelStart, bitoff)
			r := it.getNextOnLevel(level-1, count, val)
			if r != -1 {
				return r
			}
		}
		if it.isRow {
			val = it.tree.incrementNForLevel(val, 1, level)
			newoffinrun = it.tree.offsetTForLayer(it.rowcol, val, level)
		} else {
			panic("Is Column")
		}
		if newoffinrun < offInRun {
			return -1
		}
		offInRun = newoffinrun
	}
}

func (it *Iterator) getNextOnLeaf(leaflayercount, try int) int {
	leafoffset := leaflayercount * it.tree.lk.bitsPerLayer
	var bitoff, newbitoff int
	if it.isRow {
		bitoff = it.tree.offsetL(it.rowcol, try)
	} else {
		bitoff = it.tree.offsetL(try, it.rowcol)
	}
	for {
		// Test
		if it.tree.lbits.Get(leafoffset + bitoff) {
			return try
		}
		// Increment on this layer
		if it.isRow {
			try++
			newbitoff = it.tree.offsetL(it.rowcol, try)
		} else {
			panic("Is column")
		}
		// See if we've run off the edge
		if newbitoff < bitoff {
			return -1
		}
		bitoff = newbitoff
	}

}

func (it *Iterator) ExtractAll() []int {
	var out []int
	for it.Next() {
		out = append(out, it.Value())
	}
	return out
}
