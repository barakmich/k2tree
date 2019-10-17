package k2tree

import "fmt"

type RowIterator struct {
	tree   *K2Tree
	offset int
	row    int
}

func newRowIterator(tree *K2Tree, row int) *RowIterator {
	return &RowIterator{
		tree:   tree,
		offset: -1,
		row:    row,
	}
}

func (it *RowIterator) Next() bool {
	it.offset = it.getNext(it.offset)
	if it.offset != -1 {
		return true
	}
	return false
}

func (it *RowIterator) Value() int {
	return it.offset
}

func (it *RowIterator) getNext(off int) int {
	try := off + 1
	levels := it.tree.levels
	nextval := it.getNextOnLevel(levels, 0, try)
	return nextval
}

func (it *RowIterator) getNextOnLevel(level, sublayeroff, val int) int {
	// Invariant: Returned int must be >= val if the value is found or
	// -1 if the function reaches the end of the run of bits.
	fmt.Println("gnolevel", "level:", level, "sublayeroff:", sublayeroff, "val:", val)
	if level == 0 {
		return it.getNextOnLeaf(sublayeroff, val)
	}

	startRun := sublayeroff * it.tree.tk.bitsPerLayer
	levelStart := it.tree.levelOffsets[level]
	offInRun := it.tree.offsetTForLayer(it.row, val, level)
	var newoffinrun int
	fmt.Println("startrun", startRun, "levelStart", levelStart, "offInRun", offInRun)

	for {
		bitoff := levelStart + startRun + offInRun
		it.tree.printLevel(level, bitoff)
		if it.tree.tbits.Get(bitoff) {
			count := it.tree.tbits.Count(levelStart, bitoff)
			r := it.getNextOnLevel(level-1, count, val)
			if r != -1 {
				return r
			}
		}
		val = it.tree.incrementNForLevel(val, 1, level)
		newoffinrun = it.tree.offsetTForLayer(it.row, val, level)
		if newoffinrun < offInRun {
			return -1
		}
		offInRun = newoffinrun
	}
}

func (it *RowIterator) getNextOnLeaf(leaflayercount, try int) int {
	fmt.Println("gnolayer", leaflayercount, try)
	leafoffset := leaflayercount * it.tree.lk.bitsPerLayer
	bitoff := try & it.tree.lk.maskPerLayer
	for {
		// Test
		it.tree.printBase(leafoffset + bitoff)
		if it.tree.lbits.Get(leafoffset + bitoff) {
			return try
		}
		// Increment on this layer
		try++
		bitoff++
		newbitoff := bitoff & it.tree.lk.maskPerLayer
		// See if we've run off the edge
		if newbitoff < bitoff {
			return -1
		}
		bitoff = newbitoff
	}

}
