use crate::bitarray::BitArray;
use crate::error::K2TreeError;
use bitvec::prelude::*;

/// BitVecArray is a bit array backed by the bitvec crate.
/// This provides optimized bit operations using the bitvec library.
#[derive(Debug, Clone)]
pub struct BitVecArray {
    bits: BitVec<u8, Msb0>,
    total: usize,
}

impl BitVecArray {
    /// Creates a new empty BitVecArray.
    pub fn new() -> Self {
        BitVecArray {
            bits: BitVec::new(),
            total: 0,
        }
    }
}

impl Default for BitVecArray {
    fn default() -> Self {
        Self::new()
    }
}

impl BitArray for BitVecArray {
    fn len(&self) -> usize {
        self.bits.len()
    }

    fn set(&mut self, at: usize, val: bool) {
        if at >= self.bits.len() {
            panic!("can't set a bit beyond the size of the array");
        }

        let orig = self.bits[at];
        self.bits.set(at, val);

        if orig != val {
            if val {
                self.total += 1;
            } else {
                self.total -= 1;
            }
        }
    }

    fn get(&self, at: usize) -> bool {
        self.bits[at]
    }

    fn count(&self, from: usize, to: usize) -> usize {
        let (from, to) = if from > to { (to, from) } else { (from, to) };

        if from > self.bits.len() || to > self.bits.len() {
            panic!("out of range");
        }

        if from == to {
            return 0;
        }

        self.bits[from..to].count_ones()
    }

    fn total(&self) -> usize {
        self.total
    }

    fn insert(&mut self, n: usize, at: usize) -> Result<(), K2TreeError> {
        if at > self.bits.len() {
            panic!("can't extend starting at a too large offset");
        }

        if n == 0 {
            return Ok(());
        }

        // Use splice to efficiently insert n zero bits at position at
        // Create a temporary bitvec of n zeros
        let zeros = bitvec![u8, Msb0; 0; n];
        self.bits.splice(at..at, zeros.iter().by_vals());

        Ok(())
    }

    fn debug(&self) -> String {
        let mut s = format!("L{} T{} ", self.bits.len(), self.total);

        let limit = self.bits.len().min(160); // Show first 20 bytes (160 bits)
        for i in (0..limit).step_by(8) {
            let end = (i + 8).min(limit);
            for j in i..end {
                s.push(if self.bits[j] { '1' } else { '0' });
            }
            s.push(' ');
        }

        if limit < self.bits.len() {
            s.push_str("(first 20 bytes)");
        }

        s
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_bitvec_array_basic() {
        let mut arr = BitVecArray::new();
        assert_eq!(arr.len(), 0);
        assert_eq!(arr.total(), 0);

        // Insert some bits
        arr.insert(8, 0).unwrap();
        assert_eq!(arr.len(), 8);
        assert_eq!(arr.total(), 0);

        arr.set(0, true);
        assert_eq!(arr.get(0), true);
        assert_eq!(arr.total(), 1);

        arr.set(7, true);
        assert_eq!(arr.get(7), true);
        assert_eq!(arr.total(), 2);
    }

    #[test]
    fn test_bitvec_array_count() {
        let mut arr = BitVecArray::new();
        arr.insert(16, 0).unwrap();

        arr.set(0, true);
        arr.set(3, true);
        arr.set(7, true);
        arr.set(8, true);

        assert_eq!(arr.count(0, 8), 3);
        assert_eq!(arr.count(0, 16), 4);
        assert_eq!(arr.count(8, 16), 1);
    }

    #[test]
    fn test_bitvec_array_insert_middle() {
        let mut arr = BitVecArray::new();
        arr.insert(8, 0).unwrap();

        // Set some bits: 10101010
        arr.set(0, true);
        arr.set(2, true);
        arr.set(4, true);
        arr.set(6, true);

        // Insert 4 bits at position 4
        arr.insert(4, 4).unwrap();
        assert_eq!(arr.len(), 12);

        // Original pattern should be: 1010 0000 1010
        assert_eq!(arr.get(0), true);
        assert_eq!(arr.get(1), false);
        assert_eq!(arr.get(2), true);
        assert_eq!(arr.get(3), false);
        // Inserted zeros
        assert_eq!(arr.get(4), false);
        assert_eq!(arr.get(5), false);
        assert_eq!(arr.get(6), false);
        assert_eq!(arr.get(7), false);
        // Shifted bits
        assert_eq!(arr.get(8), true);
        assert_eq!(arr.get(9), false);
        assert_eq!(arr.get(10), true);
        assert_eq!(arr.get(11), false);
    }
}
