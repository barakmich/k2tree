use crate::bitarray::{insert_four_bits, BitArray};
use crate::error::K2TreeError;

/// Fast popcount using u64 chunks for better performance.
/// Processes 8 bytes at a time, which typically compiles to a single POPCNT instruction.
#[inline]
fn count_ones_fast(bytes: &[u8]) -> usize {
    let mut count = 0;

    // Process 8-byte chunks as u64
    let chunks = bytes.chunks_exact(8);
    let remainder = chunks.remainder();

    for chunk in chunks {
        let value = u64::from_ne_bytes(chunk.try_into().unwrap());
        count += value.count_ones() as usize;
    }

    // Process remaining bytes
    for &byte in remainder {
        count += byte.count_ones() as usize;
    }

    count
}

/// SliceArray is a simple bit array backed by a Vec<u8>.
#[derive(Debug, Clone)]
pub struct SliceArray {
    bytes: Vec<u8>,
    length: usize,
    total: usize,
}

impl SliceArray {
    /// Creates a new empty SliceArray.
    pub fn new() -> Self {
        SliceArray {
            bytes: Vec::new(),
            length: 0,
            total: 0,
        }
    }

    /// Creates a SliceArray from existing bytes and length (for testing).
    #[cfg(test)]
    pub fn from_bytes(bytes: Vec<u8>, length: usize) -> Self {
        SliceArray {
            bytes,
            length,
            total: 0,
        }
    }

    /// Returns a reference to the bytes (for testing).
    #[cfg(test)]
    pub fn bytes(&self) -> &[u8] {
        &self.bytes
    }
}

impl Default for SliceArray {
    fn default() -> Self {
        Self::new()
    }
}

impl BitArray for SliceArray {
    fn len(&self) -> usize {
        self.length
    }

    fn set(&mut self, at: usize, val: bool) {
        if at >= self.length {
            panic!("can't set a bit beyond the size of the array");
        }
        let off = at >> 3;
        let bit = (at & 0x07) as u8;
        let mask = 0x01 << (7 - bit);
        let orig = self.bytes[off];

        if val {
            self.bytes[off] |= mask;
        } else {
            self.bytes[off] &= !mask;
        }

        if self.bytes[off] != orig {
            if val {
                self.total += 1;
            } else {
                self.total -= 1;
            }
        }
    }

    #[inline(always)]
    fn get(&self, at: usize) -> bool {
        let off = at >> 3;
        let bit = (at & 0x07) as u8;
        let mask = 0x01 << (7 - bit);
        // SAFETY: K2Tree never calls get() with an out-of-bounds index.
        // The bounds are guaranteed by the tree structure invariants.
        unsafe { (*self.bytes.get_unchecked(off) & mask) != 0x00 }
    }

    #[inline]
    fn count(&self, from: usize, to: usize) -> usize {
        let (from, to) = if from > to { (to, from) } else { (from, to) };

        if from > self.length || to > self.length {
            panic!("out of range");
        }

        if from == to {
            return 0;
        }

        let start_off = from >> 3;
        let start_bit = (from & 0x07) as u8;
        let end_off = to >> 3;
        let end_bit = (to & 0x07) as u8;

        if start_off == end_off {
            let a = 0xFF >> start_bit;
            let b = 0xFF >> end_bit;
            // SAFETY: start_off is computed from valid bit indices
            return unsafe {
                (*self.bytes.get_unchecked(start_off) & (a & !b)).count_ones() as usize
            };
        }

        let mut c = 0;
        let mut current_off = start_off;

        if start_bit != 0 {
            // SAFETY: current_off is valid as checked above
            c += unsafe {
                (*self.bytes.get_unchecked(current_off) & (0xFF >> start_bit)).count_ones() as usize
            };
            current_off += 1;
        }

        if end_bit != 0 {
            // SAFETY: end_off is valid as checked above
            c += unsafe {
                (*self.bytes.get_unchecked(end_off) & (0xFF & !(0xFF >> end_bit))).count_ones()
                    as usize
            };
        }

        // Optimized counting using u64 chunks for better performance
        // SAFETY: slice indices are valid from the checks above
        let slice = unsafe { self.bytes.get_unchecked(current_off..end_off) };
        c += count_ones_fast(slice);

        c
    }

    fn total(&self) -> usize {
        self.total
    }

    fn insert(&mut self, n: usize, at: usize) -> Result<(), K2TreeError> {
        if at > self.length {
            panic!("can't extend starting at a too large offset");
        }

        if n == 0 {
            return Ok(());
        }

        if at % 4 != 0 {
            panic!("can only insert a sliceArray at offset multiples of 4");
        }

        if n % 8 == 0 {
            self.insert_eight(n, at)?;
        } else if n == 4 {
            self.insert_four(at)?;
        } else if n % 4 == 0 {
            let mult8 = (n >> 3) << 3;
            self.insert_eight(mult8, at)?;
            self.insert_four(at)?;
        } else {
            panic!("can only extend a sliceArray by nibbles or multiples of 8");
        }

        self.length += n;
        Ok(())
    }

    fn debug(&self) -> String {
        let mut s = format!("L{} T{} ", self.length, self.total);
        for (i, byte) in self.bytes.iter().enumerate() {
            s.push_str(&format!("{:08b} ", byte));
            if i > 20 {
                s.push_str("(first 20)");
                break;
            }
        }
        s
    }
}

impl SliceArray {
    fn insert_four(&mut self, at: usize) -> Result<(), K2TreeError> {
        if self.length % 8 == 0 {
            // We need more space
            self.bytes.push(0x00);
        }

        let mut off = at >> 3;
        let mut inbyte = 0u8;

        if at % 8 != 0 {
            inbyte = self.bytes[off] << 4;
            self.bytes[off] &= 0xF0;
            off += 1;
        }

        let out_byte = insert_four_bits(&mut self.bytes[off..], inbyte);
        if out_byte != 0x00 {
            panic!("Overshot");
        }

        Ok(())
    }

    fn insert_eight(&mut self, n: usize, at: usize) -> Result<(), K2TreeError> {
        let n_bytes = n >> 3;
        let old_len = self.bytes.len();
        self.bytes.extend(vec![0u8; n_bytes]);

        if at == self.length {
            return Ok(());
        }

        let off = at >> 3;

        if at % 8 == 0 {
            // Copy from off..old_len to off+n_bytes
            self.bytes.copy_within(off..old_len, off + n_bytes);
            self.bytes[off..off + n_bytes].fill(0x00);
        } else {
            // Copy from off+1..old_len to off+1+n_bytes
            self.bytes.copy_within(off + 1..old_len, off + 1 + n_bytes);
            self.bytes[off + 1..off + 1 + n_bytes].fill(0x00);
            self.bytes[off + n_bytes] = self.bytes[off] & 0x0F;
            self.bytes[off] &= 0xF0;
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_slice_array_basic() {
        let mut arr = SliceArray::new();
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
    fn test_slice_array_count() {
        let mut arr = SliceArray::new();
        arr.insert(16, 0).unwrap();

        arr.set(0, true);
        arr.set(3, true);
        arr.set(7, true);
        arr.set(8, true);

        assert_eq!(arr.count(0, 8), 3);
        assert_eq!(arr.count(0, 16), 4);
        assert_eq!(arr.count(8, 16), 1);
    }
}
