package k2tree

// iterFrame holds the cursor state at one T level.
type iterFrame struct {
	sublayeroff int // which block at this level
	val         int // column value at this level (bits below this level are 0)
	offInRun    int // current bit within the row's block: x*kPerLayer + y
	runCount    int // Count(levelStart, levelStart+sublayeroff*bitsPerLayer+offInRun)
	baseOff     int // x * kPerLayer — row's fixed component at this T level
}

// Iterator iterates over edges in a K2Tree.
type Iterator struct {
	tree   *K2Tree
	rowcol int
	isRow  bool
	offset int // last returned column (-1 = not yet started)
	done   bool

	// tstack[i] is the frame for T level (i+1).
	// tstack[0] = level 1 (lowest T level, parent of leaf).
	// tstack[tree.levels-1] = top T level.
	tstack     []iterFrame
	leafSub    int // sublayeroff for the leaf
	leafCol    int // current column in the leaf
	leafBitoff int // offsetL(rowcol, it.offset) — bitoff of the last returned column
}

func newRowIterator(tree *K2Tree, row int) *Iterator {
	it := &Iterator{
		tree:   tree,
		rowcol: row,
		isRow:  true,
		offset: -1,
		done:   tree.levels == 0,
	}
	if tree.levels > 0 {
		it.tstack = make([]iterFrame, tree.levels)
		for i := 0; i < tree.levels; i++ {
			// baseOff = x * kPerLayer, the row's fixed component at T level i+1.
			it.tstack[i].baseOff = tree.offsetTForLayer(row, 0, i+1)
		}
	}
	return it
}

func newColumnIterator(tree *K2Tree, col int) *Iterator {
	it := &Iterator{
		tree:   tree,
		rowcol: col,
		isRow:  false,
		offset: -1,
		done:   tree.levels == 0,
	}
	if tree.levels > 0 {
		it.tstack = make([]iterFrame, tree.levels)
		for i := 0; i < tree.levels; i++ {
			// baseOff = y * 1 — column's fixed component at T level i+1.
			it.tstack[i].baseOff = tree.offsetTForLayer(0, col, i+1)
		}
	}
	return it
}

func (it *Iterator) Next() bool {
	if it.done {
		return false
	}
	if it.offset == -1 {
		// First call: descend from the top.
		if !it.descendInto(it.tree.levels-1, 0, 0) {
			it.done = true
			return false
		}
		return true
	}

	// Try to advance within the current leaf block.
	// If the next value's bitoff increases, it might be in the same block.
	// advanceLeaf handles both same-block and cross-block cases.
	nextCol := it.offset + 1
	nextBitoff := it.offsetLForVal(nextCol)
	if nextBitoff > it.leafBitoff {
		it.leafCol = nextCol
		if it.advanceLeaf() {
			return true
		}
	}

	// Leaf block exhausted (or next col is in a new block). Bubble up.
	for i := 0; i < it.tree.levels; i++ {
		frame := &it.tstack[i]
		level := i + 1
		levelStart := it.tree.levelOffsets[level]

		// The bit at frame.offInRun was SET (we descended from it).
		// Advance past it: the set bit contributes +1 to runCount.
		frame.runCount++
		newVal := it.tree.incrementNForLevel(frame.val, 1, level)
		newOffInRun := it.offsetTForVal(newVal, level)
		if newOffInRun < frame.offInRun {
			// Row's block at this level is exhausted; bubble up further.
			continue
		}
		frame.val = newVal
		frame.offInRun = newOffInRun

		// Scan-and-descend loop: handles dead ends (subtrees with bits for
		// other rows but none for it.rowcol) by advancing within this level.
		for {
			if !it.scanFrame(i) {
				break // this level's block is exhausted; bubble up
			}

			// Recalculate runCount after scanFrame (it may have skipped set bits).
			bitoff := levelStart + frame.sublayeroff*it.tree.tk.bitsPerLayer + frame.offInRun
			frame.runCount = it.tree.tbits.Count(levelStart, bitoff)

			if it.descendInto(i-1, frame.runCount, frame.val) {
				return true
			}

			// descendInto returned false: no result for this row in that subtree.
			// Advance past the current set bit and retry at this level.
			frame.runCount++
			newVal2 := it.tree.incrementNForLevel(frame.val, 1, level)
			newOffInRun2 := it.offsetTForVal(newVal2, level)
			if newOffInRun2 < frame.offInRun {
				break // this level's block is now exhausted
			}
			frame.val = newVal2
			frame.offInRun = newOffInRun2
		}
	}

	it.done = true
	return false
}

// offsetTForVal returns the offset of (rowcol, val) or (val, rowcol) in a T block.
func (it *Iterator) offsetTForVal(val, level int) int {
	if it.isRow {
		return it.tree.offsetTForLayer(it.rowcol, val, level)
	}
	return it.tree.offsetTForLayer(val, it.rowcol, level)
}

// offsetLForVal returns the leaf offset for the current (rowcol, val) pair.
func (it *Iterator) offsetLForVal(v int) int {
	if it.isRow {
		return it.tree.offsetL(it.rowcol, v)
	}
	return it.tree.offsetL(v, it.rowcol)
}

// scanFrame scans tstack[i] from frame.offInRun for the next set bit.
// On each unset bit it advances frame.val and frame.offInRun; runCount is
// NOT modified (unset bits don't change it — the caller increments runCount
// when moving past a set bit).
// Returns true when a set bit is found at the updated frame.offInRun.
func (it *Iterator) scanFrame(i int) bool {
	frame := &it.tstack[i]
	level := i + 1
	levelStart := it.tree.levelOffsets[level]
	for {
		bitoff := levelStart + frame.sublayeroff*it.tree.tk.bitsPerLayer + frame.offInRun
		if it.tree.tbits.Get(bitoff) {
			return true
		}
		// Unset: advance val/offInRun; runCount stays the same.
		newVal := it.tree.incrementNForLevel(frame.val, 1, level)
		newOffInRun := it.offsetTForVal(newVal, level)
		if newOffInRun < frame.offInRun {
			// Wrapped: block exhausted.
			frame.val = newVal
			frame.offInRun = newOffInRun
			return false
		}
		frame.val = newVal
		frame.offInRun = newOffInRun
	}
}

// descendInto initialises tstack[startIdx] down to tstack[0], then the leaf.
// sublayeroff is the child-block index for the block at tstack[startIdx].
// val is the starting column (bits for levels ≤ startIdx are 0).
// Returns true if a value was found and stored in it.offset.
//
// Dead ends — where a T-level bit is set by another row sharing the same
// x-path but diverging at a lower level — are handled by retrying within
// this level before returning false.
func (it *Iterator) descendInto(startIdx int, sublayeroff int, val int) bool {
	if startIdx < 0 {
		it.leafSub = sublayeroff
		it.leafCol = val
		return it.advanceLeaf()
	}

	frame := &it.tstack[startIdx]
	level := startIdx + 1
	levelStart := it.tree.levelOffsets[level]

	frame.sublayeroff = sublayeroff
	frame.val = val
	frame.offInRun = frame.baseOff

	// Find the first set bit; one Count call to initialise runCount.
	if !it.scanFrame(startIdx) {
		return false
	}
	bitoff := levelStart + sublayeroff*it.tree.tk.bitsPerLayer + frame.offInRun
	frame.runCount = it.tree.tbits.Count(levelStart, bitoff)

	for {
		if it.descendInto(startIdx-1, frame.runCount, frame.val) {
			return true
		}

		// Sub-level returned false (dead end or genuinely empty for this row).
		// Advance past the current set bit — it contributes +1 to runCount —
		// then scan for the next set bit.
		frame.runCount++
		newVal := it.tree.incrementNForLevel(frame.val, 1, level)
		newOffInRun := it.offsetTForVal(newVal, level)
		if newOffInRun < frame.offInRun {
			return false // this level's block exhausted
		}
		frame.val = newVal
		frame.offInRun = newOffInRun

		if !it.scanFrame(startIdx) {
			return false // this level's block exhausted after scanning
		}
		// Recalculate runCount after scanFrame (it may have skipped set bits).
		bitoff2 := levelStart + frame.sublayeroff*it.tree.tk.bitsPerLayer + frame.offInRun
		frame.runCount = it.tree.tbits.Count(levelStart, bitoff2)
	}
}

// advanceLeaf scans forward from it.leafCol within the current leaf block
// (identified by it.leafSub). Updates it.leafCol, it.leafBitoff, and
// it.offset on success; returns false when the leaf block is exhausted.
func (it *Iterator) advanceLeaf() bool {
	leafoffset := it.leafSub * it.tree.lk.bitsPerLayer
	v := it.leafCol
	bitoff := it.offsetLForVal(v)
	for {
		if it.tree.lbits.Get(leafoffset + bitoff) {
			it.leafCol = v
			it.leafBitoff = bitoff
			it.offset = v
			return true
		}
		v++
		newbitoff := it.offsetLForVal(v)
		if newbitoff < bitoff {
			it.leafCol = v
			return false // leaf block exhausted
		}
		bitoff = newbitoff
	}
}

func (it *Iterator) Value() int {
	return it.offset
}

func (it *Iterator) ExtractAll() []int {
	var out []int
	for it.Next() {
		out = append(out, it.Value())
	}
	return out
}
