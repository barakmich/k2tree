use crate::bitarray::BitArray;
use crate::error::K2TreeError;

/// QuartileIndex is a performance optimization wrapper around a BitArray.
/// It divides the array into quartiles and maintains precomputed counts,
/// allowing faster range count queries.
#[derive(Debug)]
pub struct QuartileIndex<T: BitArray> {
    bits: T,
    offsets: [usize; 3],
    counts: [usize; 3],
}

impl<T: BitArray> QuartileIndex<T> {
    /// Creates a new QuartileIndex wrapping the given BitArray.
    pub fn new(bits: T) -> Self {
        let mut q = QuartileIndex {
            bits,
            offsets: [0; 3],
            counts: [0; 3],
        };
        q.rebuild();
        q
    }

    /// Rebuilds the quartile offsets and counts from scratch.
    fn rebuild(&mut self) {
        let len = self.bits.len();
        if len == 0 {
            self.offsets = [0; 3];
            self.counts = [0; 3];
            return;
        }

        self.offsets[0] = len / 4;
        self.offsets[1] = len / 2;
        self.offsets[2] = (len / 2) + (len / 4);

        self.counts[0] = self.bits.count(0, self.offsets[0]);
        self.counts[1] = self.bits.count(0, self.offsets[1]);
        self.counts[2] = self.bits.count(0, self.offsets[2]);
    }

    /// Adjusts a single quartile after an insert operation.
    fn adjust(&mut self, index: usize, n: usize, at: usize, new_offset: usize) {
        let old_offset = self.offsets[index];

        // Sanity check: insert should not shrink the array
        assert!(new_offset >= old_offset, "Inserting shrunk the array?");

        self.offsets[index] = new_offset;

        if (n + at) < old_offset {
            // Entire span below me, adjust for loss
            self.counts[index] -= self.bits.count(new_offset, old_offset + n);
        } else if at >= old_offset {
            // Entire span above me, adjust for gain
            self.counts[index] += self.bits.count(old_offset, new_offset);
        } else {
            // Span intersects me - recalculate from scratch
            self.counts[index] = self.bits.count(0, new_offset);
        }
    }

    /// Computes the count from zero to the given value (inclusive).
    fn zero_count(&self, to: usize) -> usize {
        let mut prev_offset = 0;
        let mut prev_count = 0;

        for i in 0..3 {
            let offset = self.offsets[i];
            if to < offset {
                // Choose the faster path: count from previous or from current
                if offset - to < to - prev_offset {
                    return self.counts[i] - self.bits.count(to, offset);
                } else {
                    return self.bits.count(prev_offset, to) + prev_count;
                }
            }
            prev_offset = offset;
            prev_count = self.counts[i];
        }

        // Handle the final segment
        let len = self.bits.len();
        if len - to < to - prev_offset {
            return self.bits.total() - self.bits.count(to, len);
        } else {
            return self.bits.count(prev_offset, to) + prev_count;
        }
    }
}

impl<T: BitArray> BitArray for QuartileIndex<T> {
    fn len(&self) -> usize {
        self.bits.len()
    }

    fn set(&mut self, at: usize, val: bool) {
        let current = self.bits.get(at);
        if current == val {
            return; // No change needed
        }

        self.bits.set(at, val);

        let delta = if val { 1isize } else { -1isize };

        // Update all quartiles that are affected by this position
        for i in 0..3 {
            let offset = self.offsets[i];
            if at < offset {
                self.counts[i] = (self.counts[i] as isize + delta) as usize;
            }
        }
    }

    fn get(&self, at: usize) -> bool {
        self.bits.get(at)
    }

    fn count(&self, from: usize, to: usize) -> usize {
        if from == 0 {
            return self.zero_count(to);
        }
        self.zero_count(to) - self.zero_count(from)
    }

    fn total(&self) -> usize {
        self.bits.total()
    }

    fn insert(&mut self, n: usize, at: usize) -> Result<(), K2TreeError> {
        // QuartileIndex can only extend by nibbles (multiples of 4)
        if n % 4 != 0 {
            panic!("can only extend by nibbles (multiples of 4)");
        }

        // Delegate to the underlying bitarray
        self.bits.insert(n, at)?;

        let new_len = self.bits.len();

        // Rebuild all quartiles
        for i in 0..3 {
            let new_offset = new_len * (i + 1) / 4;
            self.adjust(i, n, at, new_offset);
        }

        Ok(())
    }

    fn debug(&self) -> String {
        format!(
            "QuartileIndex {{ offsets: {:?}, counts: {:?}, internal: {} }}",
            self.offsets,
            self.counts,
            self.bits.debug()
        )
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::bitarray::SliceArray;

    #[test]
    fn test_quartile_index_basic() {
        let mut arr = SliceArray::new();
        arr.insert(16, 0).unwrap();
        arr.set(0, true);
        arr.set(3, true);
        arr.set(7, true);
        arr.set(8, true);
        arr.set(12, true);
        arr.set(15, true);

        let quartile = QuartileIndex::new(arr);

        println!("Len: {}, Total: {}", quartile.len(), quartile.total());
        println!("Offsets: {:?}", quartile.offsets);
        println!("Counts: {:?}", quartile.counts);
        println!("Count(0, 16) = {}", quartile.count(0, 16));
        println!("Count(0, 8) = {}", quartile.count(0, 8));
        println!("Count(8, 16) = {}", quartile.count(8, 16));

        assert_eq!(quartile.len(), 16);
        assert_eq!(quartile.total(), 6);
        assert_eq!(quartile.count(0, 16), 6);
        assert_eq!(quartile.count(0, 8), 3);
        assert_eq!(quartile.count(8, 16), 3);
    }

    #[test]
    fn test_quartile_index_count() {
        let mut arr = SliceArray::new();
        arr.insert(32, 0).unwrap();

        // Set some bits at known positions
        for i in [0, 3, 7, 10, 15, 16, 20, 24, 28, 31].iter() {
            arr.set(*i, true);
        }

        let quartile = QuartileIndex::new(arr.clone());

        // Test various ranges - these should match the underlying SliceArray
        assert_eq!(quartile.count(0, 8), arr.count(0, 8));
        assert_eq!(quartile.count(8, 16), arr.count(8, 16));
        assert_eq!(quartile.count(16, 24), arr.count(16, 24));
        assert_eq!(quartile.count(24, 32), arr.count(24, 32));
        assert_eq!(quartile.count(0, 32), arr.count(0, 32));
    }

    #[test]
    fn test_quartile_index_set() {
        let mut arr = SliceArray::new();
        arr.insert(16, 0).unwrap();

        let mut quartile = QuartileIndex::new(arr);

        // Set a bit
        quartile.set(5, true);
        assert_eq!(quartile.count(0, 16), 1);

        // Unset a bit (should do nothing since it's not set)
        quartile.set(5, true);
        assert_eq!(quartile.count(0, 16), 1);

        // Set another bit
        quartile.set(10, true);
        assert_eq!(quartile.count(0, 16), 2);

        // Unset the first bit
        quartile.set(5, false);
        assert_eq!(quartile.count(0, 16), 1);
    }

    #[test]
    fn test_quartile_index_invariants() {
        // This test is ported from the Go test TestQuartileCount
        // It tests that the count invariants are maintained after insertions
        let mut quartile = QuartileIndex::new(SliceArray::new());
        quartile.insert(24, 0).unwrap();
        quartile.set(3, true);
        quartile.insert(8, 0).unwrap();

        // Check invariants: each count should match the actual count up to that offset
        for i in 0..3 {
            let offset = quartile.offsets[i];
            let expected = quartile.bits.count(0, offset);
            assert_eq!(quartile.counts[i], expected,
                "Count invariant failed: quartile index {}, count {}, expected {}",
                i, quartile.counts[i], expected);
        }
    }

    #[test]
    fn test_quartile_index_insert() {
        let mut arr = SliceArray::new();
        arr.insert(16, 0).unwrap();
        arr.set(0, true);
        arr.set(7, true);

        let mut quartile = QuartileIndex::new(arr);
        assert_eq!(quartile.count(0, 16), 2);

        // Insert 4 bits at position 8
        quartile.insert(4, 8).unwrap();
        assert_eq!(quartile.len(), 20);

        // Original bits should still be there
        assert_eq!(quartile.get(0), true);
        assert_eq!(quartile.get(7), true); // Was at position 7, still at 7
        assert_eq!(quartile.count(0, 20), 2);
    }
}
