use crate::bitarray::SliceArray;
use crate::k2tree::{DEFAULT_CONFIG, K2Tree};

#[cfg(test)]
mod tests {
    use super::*;

    fn simple_load(k: &mut K2Tree<SliceArray>) {
        k.add(20, 41).unwrap();
        k.add(14, 20).unwrap();
        k.add(20, 2).unwrap();
        k.add(20, 1).unwrap();
        k.add(1, 14).unwrap();
        k.add(20, 14).unwrap();
        k.add(20, 30).unwrap();
        k.add(30, 30).unwrap();
        k.add(20, 17).unwrap();
        k.add(41, 17).unwrap();
        k.add(41, 1).unwrap();
        k.add(41, 30).unwrap();
    }

    #[test]
    fn test_row_iterator() {
        struct TestCase {
            row: usize,
            expected: Vec<usize>,
        }

        let test_cases = vec![
            TestCase {
                row: 20,
                expected: vec![1, 2, 14, 17, 30, 41],
            },
            TestCase {
                row: 41,
                expected: vec![1, 17, 30],
            },
        ];

        for (i, test) in test_cases.into_iter().enumerate() {
            let mut k2 =
                K2Tree::new_with_config(SliceArray::new(), SliceArray::new(), DEFAULT_CONFIG);
            simple_load(&mut k2);

            let mut out = k2.from(test.row).extract_all();
            out.sort();

            assert_eq!(
                out.len(),
                test.expected.len(),
                "instance {} mismatch in length: out: {:?} expected {:?}",
                i,
                out,
                test.expected
            );

            for j in 0..test.expected.len() {
                assert_eq!(
                    out[j], test.expected[j],
                    "instance {} mismatch: out: {:?} expected: {:?}",
                    i, out, test.expected
                );
            }
        }
    }

    #[test]
    fn test_iterator_next_value() {
        let mut k2 = K2Tree::new_with_config(SliceArray::new(), SliceArray::new(), DEFAULT_CONFIG);
        simple_load(&mut k2);

        let mut it = k2.from(20);
        let mut values = Vec::new();

        while it.next_edge() {
            values.push(it.value());
        }

        values.sort();
        let expected = vec![1, 2, 14, 17, 30, 41];

        assert_eq!(values, expected);
    }
}
