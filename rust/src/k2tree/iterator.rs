use super::K2Tree;
use crate::bitarray::BitArray;

/// K2TreeIterator iterates over edges in a K2Tree.
pub struct K2TreeIterator<'a, T: BitArray> {
    tree: &'a K2Tree<T>,
    offset: isize,
    rowcol: usize,
    is_row: bool,
}

impl<'a, T: BitArray> K2TreeIterator<'a, T> {
    /// Creates a new row iterator (outgoing edges from node row).
    pub fn new_row(tree: &'a K2Tree<T>, row: usize) -> Self {
        K2TreeIterator {
            tree,
            offset: -1,
            rowcol: row,
            is_row: true,
        }
    }

    /// Creates a new column iterator (incoming edges to node col).
    pub fn new_column(tree: &'a K2Tree<T>, col: usize) -> Self {
        K2TreeIterator {
            tree,
            offset: -1,
            rowcol: col,
            is_row: false,
        }
    }

    /// Advances the iterator to the next edge.
    /// Returns true if there is a next value, false if iteration is complete.
    pub fn next_edge(&mut self) -> bool {
        self.offset = self.get_next(self.offset);
        self.offset != -1
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

    #[inline]
    fn get_next(&self, off: isize) -> isize {
        let try_val = off + 1;
        let levels = self.tree.levels;
        self.get_next_on_level(levels, 0, try_val)
    }

    #[inline]
    fn get_next_on_level(&self, level: usize, sublayeroff: usize, val: isize) -> isize {
        // Invariant: Returned int must be >= val if the value is found or
        // -1 if the function reaches the end of the run of bits.
        if level == 0 {
            return self.get_next_on_leaf(sublayeroff, val);
        }

        let start_run = sublayeroff * self.tree.tk.bits_per_layer;
        let level_start = self.tree.level_offsets[level];
        let mut val = val;
        let mut off_in_run = if self.is_row {
            self.tree
                .offset_t_for_layer(self.rowcol, val as usize, level)
        } else {
            self.tree
                .offset_t_for_layer(val as usize, self.rowcol, level)
        };

        loop {
            let bitoff = level_start + start_run + off_in_run;
            if self.tree.tbits.get(bitoff) {
                let count = self.tree.tbits.count(level_start, bitoff);
                let r = self.get_next_on_level(level - 1, count, val);
                if r != -1 {
                    return r;
                }
            }

            if self.is_row {
                val = self.tree.increment_n_for_level(val as usize, 1, level) as isize;
                let newoffinrun = self
                    .tree
                    .offset_t_for_layer(self.rowcol, val as usize, level);
                if newoffinrun < off_in_run {
                    return -1;
                }
                off_in_run = newoffinrun;
            } else {
                panic!("Is Column");
            }
        }
    }

    #[inline]
    fn get_next_on_leaf(&self, leaflayercount: usize, try_val: isize) -> isize {
        let leafoffset = leaflayercount * self.tree.lk.bits_per_layer;
        let mut try_val = try_val;
        let mut bitoff = if self.is_row {
            self.tree.offset_l(self.rowcol, try_val as usize)
        } else {
            self.tree.offset_l(try_val as usize, self.rowcol)
        };

        loop {
            // Test
            if self.tree.lbits.get(leafoffset + bitoff) {
                return try_val;
            }

            // Increment on this layer
            if self.is_row {
                try_val += 1;
                let newbitoff = self.tree.offset_l(self.rowcol, try_val as usize);
                // See if we've run off the edge
                if newbitoff < bitoff {
                    return -1;
                }
                bitoff = newbitoff;
            } else {
                panic!("Is column");
            }
        }
    }
}
