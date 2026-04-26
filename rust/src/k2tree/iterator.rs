use super::K2Tree;
use crate::bitarray::BitArray;

struct IterFrame {
    sublayeroff: usize,
    val: usize,
    off_in_run: usize,
    run_count: usize,
    base_off: usize,
}

/// K2TreeIterator iterates over edges in a K2Tree.
pub struct K2TreeIterator<'a, T: BitArray> {
    tree: &'a K2Tree<T>,
    rowcol: usize,
    is_row: bool,
    offset: isize,
    done: bool,
    tstack: Vec<IterFrame>,
    leaf_sub: usize,
    leaf_col: usize,
    leaf_bitoff: usize,
}

impl<'a, T: BitArray> K2TreeIterator<'a, T> {
    /// Creates a new row iterator (outgoing edges from node row).
    pub fn new_row(tree: &'a K2Tree<T>, row: usize) -> Self {
        let done = tree.levels == 0;
        let tstack = if tree.levels > 0 {
            (0..tree.levels)
                .map(|i| IterFrame {
                    sublayeroff: 0,
                    val: 0,
                    off_in_run: 0,
                    run_count: 0,
                    base_off: tree.offset_t_for_layer(row, 0, i + 1),
                })
                .collect()
        } else {
            Vec::new()
        };
        K2TreeIterator {
            tree,
            rowcol: row,
            is_row: true,
            offset: -1,
            done,
            tstack,
            leaf_sub: 0,
            leaf_col: 0,
            leaf_bitoff: 0,
        }
    }

    /// Creates a new column iterator (incoming edges to node col).
    pub fn new_column(tree: &'a K2Tree<T>, col: usize) -> Self {
        let done = tree.levels == 0;
        let tstack = if tree.levels > 0 {
            (0..tree.levels)
                .map(|i| IterFrame {
                    sublayeroff: 0,
                    val: 0,
                    off_in_run: 0,
                    run_count: 0,
                    base_off: tree.offset_t_for_layer(0, col, i + 1),
                })
                .collect()
        } else {
            Vec::new()
        };
        K2TreeIterator {
            tree,
            rowcol: col,
            is_row: false,
            offset: -1,
            done,
            tstack,
            leaf_sub: 0,
            leaf_col: 0,
            leaf_bitoff: 0,
        }
    }

    /// Like `offset_t_for_layer` but swaps arguments for column iteration.
    fn offset_t_for_val(&self, val: usize, level: usize) -> usize {
        if self.is_row {
            self.tree.offset_t_for_layer(self.rowcol, val, level)
        } else {
            self.tree.offset_t_for_layer(val, self.rowcol, level)
        }
    }

    /// Like `offset_l` but swaps arguments for column iteration.
    fn offset_l_for_val(&self, val: usize) -> usize {
        if self.is_row {
            self.tree.offset_l(self.rowcol, val)
        } else {
            self.tree.offset_l(val, self.rowcol)
        }
    }

    /// Advances the iterator to the next edge.
    /// Returns true if there is a next value, false if iteration is complete.
    pub fn next_edge(&mut self) -> bool {
        if self.done {
            return false;
        }
        if self.offset == -1 {
            if !self.descend_into(self.tree.levels as isize - 1, 0, 0) {
                self.done = true;
                return false;
            }
            return true;
        }

        // Try to advance within the current leaf block.
        let next_col = (self.offset + 1) as usize;
        let next_bitoff = self.offset_l_for_val(next_col);
        if next_bitoff > self.leaf_bitoff {
            self.leaf_col = next_col;
            if self.advance_leaf() {
                return true;
            }
        }

        // Leaf block exhausted. Bubble up through T levels.
        for i in 0..self.tree.levels {
            let level = i + 1;
            let level_start = self.tree.level_offsets[level];

            self.tstack[i].run_count += 1;
            let new_val = self.tree.increment_n_for_level(self.tstack[i].val, 1, level);
            let new_off_in_run = self.offset_t_for_val(new_val, level);
            if new_off_in_run < self.tstack[i].off_in_run {
                continue;
            }
            self.tstack[i].val = new_val;
            self.tstack[i].off_in_run = new_off_in_run;

            loop {
                if !self.scan_frame(i) {
                    break;
                }

                // Recalculate run_count after scan_frame (it may have skipped set bits).
                let bitoff = level_start
                    + self.tstack[i].sublayeroff * self.tree.tk.bits_per_layer
                    + self.tstack[i].off_in_run;
                self.tstack[i].run_count = self.tree.tbits.count(level_start, bitoff);

                let run_count = self.tstack[i].run_count;
                let val = self.tstack[i].val;
                if self.descend_into(i as isize - 1, run_count, val) {
                    return true;
                }

                self.tstack[i].run_count += 1;
                let new_val2 = self.tree.increment_n_for_level(self.tstack[i].val, 1, level);
                let new_off2 = self.offset_t_for_val(new_val2, level);
                if new_off2 < self.tstack[i].off_in_run {
                    break;
                }
                self.tstack[i].val = new_val2;
                self.tstack[i].off_in_run = new_off2;
            }
        }

        self.done = true;
        false
    }

    /// Returns the current value (the node index of the current edge).
    pub fn value(&self) -> usize {
        self.offset as usize
    }

    /// Extracts all remaining values into a Vec.
    pub fn extract_all(&mut self) -> Vec<usize> {
        let mut out = Vec::new();
        while self.next_edge() {
            out.push(self.value());
        }
        out
    }

    // scan_frame scans tstack[i] from off_in_run for the next set bit.
    // Advances val and off_in_run on each unset bit; run_count is NOT modified.
    // Returns true when a set bit is found.
    fn scan_frame(&mut self, i: usize) -> bool {
        let level = i + 1;
        let level_start = self.tree.level_offsets[level];
        loop {
            let bitoff = level_start
                + self.tstack[i].sublayeroff * self.tree.tk.bits_per_layer
                + self.tstack[i].off_in_run;
            if self.tree.tbits.get(bitoff) {
                return true;
            }
            let new_val = self.tree.increment_n_for_level(self.tstack[i].val, 1, level);
            let new_off = self.offset_t_for_val(new_val, level);
            if new_off < self.tstack[i].off_in_run {
                self.tstack[i].val = new_val;
                self.tstack[i].off_in_run = new_off;
                return false;
            }
            self.tstack[i].val = new_val;
            self.tstack[i].off_in_run = new_off;
        }
    }

    // descend_into initialises tstack[start_idx] down to tstack[0], then the leaf.
    // Returns true if a value was found and stored in self.offset.
    fn descend_into(&mut self, start_idx: isize, sublayeroff: usize, val: usize) -> bool {
        if start_idx < 0 {
            self.leaf_sub = sublayeroff;
            self.leaf_col = val;
            return self.advance_leaf();
        }

        let i = start_idx as usize;
        let level = i + 1;
        let level_start = self.tree.level_offsets[level];

        self.tstack[i].sublayeroff = sublayeroff;
        self.tstack[i].val = val;
        self.tstack[i].off_in_run = self.tstack[i].base_off;

        if !self.scan_frame(i) {
            return false;
        }

        let bitoff = level_start
            + sublayeroff * self.tree.tk.bits_per_layer
            + self.tstack[i].off_in_run;
        self.tstack[i].run_count = self.tree.tbits.count(level_start, bitoff);

        loop {
            let run_count = self.tstack[i].run_count;
            let val = self.tstack[i].val;
            if self.descend_into(start_idx - 1, run_count, val) {
                return true;
            }

            self.tstack[i].run_count += 1;
            let new_val = self.tree.increment_n_for_level(self.tstack[i].val, 1, level);
            let new_off = self.offset_t_for_val(new_val, level);
            if new_off < self.tstack[i].off_in_run {
                return false;
            }
            self.tstack[i].val = new_val;
            self.tstack[i].off_in_run = new_off;

            if !self.scan_frame(i) {
                return false;
            }
            // Recalculate run_count after scan_frame (it may have skipped set bits).
            let bitoff2 = level_start
                + self.tstack[i].sublayeroff * self.tree.tk.bits_per_layer
                + self.tstack[i].off_in_run;
            self.tstack[i].run_count = self.tree.tbits.count(level_start, bitoff2);
        }
    }

    // advance_leaf scans forward from leaf_col within the current leaf block.
    // Updates leaf_col, leaf_bitoff, and offset on success.
    fn advance_leaf(&mut self) -> bool {
        let leafoffset = self.leaf_sub * self.tree.lk.bits_per_layer;
        let mut val = self.leaf_col;
        let mut bitoff = self.offset_l_for_val(val);
        loop {
            if self.tree.lbits.get(leafoffset + bitoff) {
                self.leaf_col = val;
                self.leaf_bitoff = bitoff;
                self.offset = val as isize;
                return true;
            }
            val += 1;
            let new_bitoff = self.offset_l_for_val(val);
            if new_bitoff < bitoff {
                self.leaf_col = val;
                return false;
            }
            bitoff = new_bitoff;
        }
    }
}
