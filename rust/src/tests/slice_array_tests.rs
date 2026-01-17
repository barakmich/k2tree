use crate::bitarray::{BitArray, SliceArray};

#[cfg(test)]
mod tests {
    use super::*;

    struct TestCase {
        n: usize,
        at: usize,
        input: Vec<u8>,
        output: Vec<u8>,
        length: usize,
    }

    #[test]
    fn test_slice_array_insert_table() {
        let test_cases = vec![
            TestCase {
                n: 4,
                at: 12,
                input: vec![0xAB, 0xCD, 0xEF],
                length: 24,
                output: vec![0xAB, 0xC0, 0xDE, 0xF0],
            },
            TestCase {
                n: 12,
                at: 4,
                input: vec![0xAB, 0xCD, 0xEF],
                length: 24,
                output: vec![0xA0, 0x00, 0xBC, 0xDE, 0xF0],
            },
            TestCase {
                n: 8,
                at: 4,
                input: vec![0xAB, 0xCD, 0xEF],
                length: 24,
                output: vec![0xA0, 0x0B, 0xCD, 0xEF],
            },
            TestCase {
                n: 16,
                at: 8,
                input: vec![0xAB, 0xCD, 0xEF],
                length: 24,
                output: vec![0xAB, 0x00, 0x00, 0xCD, 0xEF],
            },
            TestCase {
                n: 12,
                at: 8,
                input: vec![0xAB, 0xCD, 0xEF],
                length: 24,
                output: vec![0xAB, 0x00, 0x0C, 0xDE, 0xF0],
            },
            TestCase {
                n: 4,
                at: 8,
                input: vec![0xAB, 0xCD, 0xEF],
                length: 24,
                output: vec![0xAB, 0x0C, 0xDE, 0xF0],
            },
            TestCase {
                n: 4,
                at: 12,
                input: vec![0xAB, 0xCD, 0xEF, 0x10],
                length: 28,
                output: vec![0xAB, 0xC0, 0xDE, 0xF1],
            },
            TestCase {
                n: 4,
                at: 16,
                input: vec![0xAB, 0xCD, 0xEF, 0x10],
                length: 28,
                output: vec![0xAB, 0xCD, 0x0E, 0xF1],
            },
            TestCase {
                n: 12,
                at: 8,
                input: vec![0xAB, 0xCD, 0xEF, 0x10],
                length: 28,
                output: vec![0xAB, 0x00, 0x0C, 0xDE, 0xF1],
            },
            TestCase {
                n: 12,
                at: 12,
                input: vec![0xAB, 0xCD, 0xEF, 0x10],
                length: 28,
                output: vec![0xAB, 0xC0, 0x00, 0xDE, 0xF1],
            },
            TestCase {
                n: 4,
                at: 4,
                input: vec![0x19],
                length: 8,
                output: vec![0x10, 0x90],
            },
        ];

        for (i, test) in test_cases.into_iter().enumerate() {
            let mut arr = SliceArray::from_bytes(test.input.clone(), test.length);

            arr.insert(test.n, test.at)
                .unwrap_or_else(|_| panic!("insert failed for test case {}", i));

            assert_eq!(
                arr.bytes(), test.output,
                "Test case {} failed: got {:?}, expected {:?} (n: {}, at: {}, len: {})",
                i, arr.bytes(), test.output, test.n, test.at, test.length
            );
        }
    }
}
