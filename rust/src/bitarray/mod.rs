mod insert_four;
mod lru_array;
mod quartile_index;
mod slice_array;

pub use insert_four::insert_four_bits;
pub use lru_array::LruArray;
pub use quartile_index::QuartileIndex;
pub use slice_array::SliceArray;

use crate::error::K2TreeError;

/// BitArray trait defines the interface for bit array implementations.
pub trait BitArray {
    /// Returns the number of bits in the bitarray.
    fn len(&self) -> usize;

    /// Returns true if the bitarray is empty.
    fn is_empty(&self) -> bool {
        self.len() == 0
    }

    /// Sets the bit at index `at` to the value `val`.
    fn set(&mut self, at: usize, val: bool);

    /// Returns the value stored at index `at`.
    fn get(&self, at: usize) -> bool;

    /// Returns the number of set bits in the interval [from, to).
    fn count(&self, from: usize, to: usize) -> usize;

    /// Returns the total number of set bits in the entire array.
    fn total(&self) -> usize;

    /// Insert extends the bitarray by `n` bits. The bits are zeroed
    /// and start at index `at`.
    ///
    /// Example:
    /// Initial string: 11101
    /// Insert(3, 2)
    /// Resulting string: 11000101
    fn insert(&mut self, n: usize, at: usize) -> Result<(), K2TreeError>;

    /// Returns a debug string representation of the bitarray.
    fn debug(&self) -> String;
}
