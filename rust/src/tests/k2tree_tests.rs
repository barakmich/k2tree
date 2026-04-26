use crate::bitarray::{BitArray, SliceArray};
use crate::k2tree::{K2Tree, SIXTEEN_SIXTEEN_CONFIG};

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_simple_add() {
        let mut k = K2Tree::new(SliceArray::new(), SliceArray::new());
        let mut kk = K2Tree::new(SliceArray::new(), SliceArray::new());

        // Add edges 0-7 in reverse order for k
        for x in (0..8).rev() {
            k.add(x, x).unwrap();
        }

        // Add edges 0-7 in forward order for kk
        for x in 0..8 {
            kk.add(x, x).unwrap();
        }

        // Verify that both trees have the same structure
        assert_eq!(
            k.tbits().len(),
            kk.tbits().len(),
            "lengths don't match in T"
        );

        for i in 0..k.tbits().len() {
            assert_eq!(
                k.tbits().get(i),
                kk.tbits().get(i),
                "index {} doesn't match in T",
                i
            );
        }

        assert_eq!(
            k.lbits().len(),
            kk.lbits().len(),
            "lengths don't match in L"
        );

        for i in 0..k.lbits().len() {
            assert_eq!(
                k.lbits().get(i),
                kk.lbits().get(i),
                "index {} doesn't match in L",
                i
            );
        }
    }

    #[test]
    fn test_sixteen_bpl_simplified() {
        // Simplified version of TestSixteenBPL - adds fewer edges for faster testing
        let mut kk =
            K2Tree::new_with_config(SliceArray::new(), SliceArray::new(), SIXTEEN_SIXTEEN_CONFIG);

        let base = 5000000;
        // Add 1000 edges instead of 1M for faster testing
        for x in 0..1000 {
            kk.add(base + x, base + x).unwrap();
        }

        let stats = kk.stats();
        println!("{}", stats);

        // Verify that edges were added
        assert!(stats.links >= 1000, "Expected at least 1000 links");
    }

    #[test]
    fn test_basic_add_and_query() {
        let mut k = K2Tree::new(SliceArray::new(), SliceArray::new());

        // Add some edges
        k.add(0, 1).unwrap();
        k.add(0, 2).unwrap();
        k.add(1, 3).unwrap();
        k.add(2, 3).unwrap();

        // Query edges from node 0
        let edges_from_0 = k.from(0).extract_all();
        assert_eq!(edges_from_0.len(), 2);
        assert!(edges_from_0.contains(&1));
        assert!(edges_from_0.contains(&2));

        // Query edges from node 1
        let edges_from_1 = k.from(1).extract_all();
        assert_eq!(edges_from_1.len(), 1);
        assert!(edges_from_1.contains(&3));

        // Query edges from node 2
        let edges_from_2 = k.from(2).extract_all();
        assert_eq!(edges_from_2.len(), 1);
        assert!(edges_from_2.contains(&3));
    }
}
