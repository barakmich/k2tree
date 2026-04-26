mod config;
mod grow;
mod iterator;

pub use config::{
    Config, LayerDef, DEFAULT_CONFIG, FOUR_BITS_PER_LAYER, FOUR_FOUR_CONFIG,
    SIXTEEN_BITS_PER_LAYER, SIXTEEN_FOUR_CONFIG, SIXTEEN_SIXTEEN_CONFIG,
};
pub use iterator::K2TreeIterator;

use crate::bitarray::BitArray;
use crate::error::K2TreeError;
use std::fmt;

/// K2Tree is the main data structure for this package. It represents a compressed
/// representation of a graph adjacency matrix.
pub struct K2Tree<T: BitArray> {
    tbits: T,
    lbits: T,
    tk: LayerDef,
    lk: LayerDef,
    levels: usize,
    level_offsets: Vec<usize>,
}

impl<T: BitArray> K2Tree<T> {
    /// Creates a new K2Tree with the default configuration.
    pub fn new(tbits: T, lbits: T) -> Self {
        Self::new_with_config(tbits, lbits, DEFAULT_CONFIG)
    }

    /// Creates a new K2Tree with the specified configuration.
    pub fn new_with_config(tbits: T, lbits: T, config: Config) -> Self {
        K2Tree {
            tbits,
            lbits,
            tk: config.tree_layer_def,
            lk: config.cell_layer_def,
            levels: 0,
            level_offsets: Vec::new(),
        }
    }

    /// Returns the largest node index representable by this K2Tree.
    pub fn max_index(&self) -> usize {
        if self.levels == 0 {
            return 0;
        }
        self.tk.k_per_layer.pow(self.levels as u32) * self.lk.k_per_layer
    }

    /// Adds an edge from node i to node j.
    /// i and j are zero-indexed. The tree will grow to support them if necessary.
    pub fn add(&mut self, i: usize, j: usize) -> Result<(), K2TreeError> {
        if self.tbits.len() == 0 {
            self.init_tree(i.max(j))?;
        } else if i >= self.max_index() || j >= self.max_index() {
            self.grow_tree(i.max(j))?;
        }
        self.add_internal(i, j)
    }

    /// Returns an iterator over outgoing edges from node i (row iterator).
    pub fn from(&self, i: usize) -> K2TreeIterator<'_, T> {
        K2TreeIterator::new_row(self, i)
    }

    /// Returns an iterator over incoming edges to node j (column iterator).
    pub fn to(&self, j: usize) -> K2TreeIterator<'_, T> {
        K2TreeIterator::new_column(self, j)
    }

    /// Returns a reference to the tree bits (for testing).
    #[cfg(test)]
    pub fn tbits(&self) -> &T {
        &self.tbits
    }

    /// Returns a reference to the leaf bits (for testing).
    #[cfg(test)]
    pub fn lbits(&self) -> &T {
        &self.lbits
    }

    /// Returns statistics about the K2Tree's memory usage.
    pub fn stats(&self) -> Stats {
        let c = self.lbits.total();
        let bytes = (self.lbits.len() + self.tbits.len()) >> 3;
        Stats {
            bits_per_link: if c > 0 {
                (self.lbits.len() + self.tbits.len()) as f64 / c as f64
            } else {
                0.0
            },
            links: c,
            level_offsets: self.level_offsets.clone(),
            bytes,
            t_debug: self.tbits.debug(),
            l_debug: self.lbits.debug(),
        }
    }

    /// Internal helper to add an edge at (i, j).
    fn add_internal(&mut self, i: usize, j: usize) -> Result<(), K2TreeError> {
        let mut level = self.levels;
        if self.level_offsets[level] != 0 {
            panic!("top level is not offset 0?");
        }

        let mut level_offset = 0;
        let mut count = 0;

        while level != 0 {
            let level_start = self.level_offsets[level];
            let offset = self.offset_t_for_layer(i, j, level);
            let bitoff = level_start + level_offset + offset;
            count = self.tbits.count(level_start, bitoff);

            if self.tbits.get(bitoff) {
                level_offset = count * self.tk.bits_per_layer;
            } else {
                self.tbits.set(bitoff, true);
                self.insert_to_layer(level - 1, count)?;
                level_offset = count * self.tk.bits_per_layer;
            }
            level -= 1;
        }

        let offset = self.offset_l(i, j);
        let bitoff = (count * self.lk.bits_per_layer) + offset;
        self.lbits.set(bitoff, true);
        Ok(())
    }

    /// Returns the offset of (i, j) in layer l.
    #[inline(always)]
    fn offset_t_for_layer(&self, i: usize, j: usize, l: usize) -> usize {
        let spl = (l - 1) * self.tk.shift_per_layer + self.lk.shift_per_layer;
        let x = (i & (self.tk.mask_per_layer << spl)) >> spl;
        let y = (j & (self.tk.mask_per_layer << spl)) >> spl;
        (x * self.tk.k_per_layer) + y
    }

    /// Increments n by amt at the appropriate level.
    #[inline(always)]
    fn increment_n_for_level(&self, n: usize, amt: usize, l: usize) -> usize {
        let spl = (l - 1) * self.tk.shift_per_layer + self.lk.shift_per_layer;
        ((n >> spl) + amt) << spl
    }

    /// Returns the suboffset within the leaf layer for (i, j).
    #[inline(always)]
    fn offset_l(&self, i: usize, j: usize) -> usize {
        ((i & self.lk.mask_per_layer) * self.lk.k_per_layer) + (j & self.lk.mask_per_layer)
    }

    /// Computes the number of layers necessary to represent index i.
    fn necessary_layer(&self, i: usize) -> usize {
        let mut n = self.lk.k_per_layer * self.tk.k_per_layer;
        let mut l = 1;
        while n <= i {
            n *= self.tk.k_per_layer;
            l += 1;
        }
        l
    }

    // Methods from grow.rs are implemented in the grow module
    fn init_tree(&mut self, size: usize) -> Result<(), K2TreeError> {
        self::grow::init_tree(self, size)
    }

    fn grow_tree(&mut self, size: usize) -> Result<(), K2TreeError> {
        self::grow::grow_tree(self, size)
    }

    fn insert_to_layer(&mut self, l: usize, layer_count: usize) -> Result<(), K2TreeError> {
        self::grow::insert_to_layer(self, l, layer_count)
    }
}

impl<T: BitArray> fmt::Debug for K2Tree<T> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "T: {}\nL: {}\nOffsets: {:?}, Levels: {}",
            self.tbits.debug(),
            self.lbits.debug(),
            self.level_offsets,
            self.levels
        )
    }
}

/// Stats describes the memory usage of the K2Tree.
#[derive(Debug, Clone)]
pub struct Stats {
    pub bits_per_link: f64,
    pub links: usize,
    pub level_offsets: Vec<usize>,
    pub bytes: usize,
    pub t_debug: String,
    pub l_debug: String,
}

impl fmt::Display for Stats {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "\nBits Per Link: {}\nLinks: {}\nLevelOffsets: {:?}\nBytes: {}\nTDebug: {}\nLDebug: {}\n",
            self.bits_per_link,
            self.links,
            self.level_offsets,
            self.bytes,
            self.t_debug,
            self.l_debug
        )
    }
}
