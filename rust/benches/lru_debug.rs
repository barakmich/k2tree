use k2tree::{BitArray, K2Tree, LruArray, SliceArray, SIXTEEN_FOUR_CONFIG};
use rand::Rng;

fn populate_incremental_tree<T: BitArray>(n_links: usize, k2: &mut K2Tree<T>) {
    let mut rng = rand::thread_rng();
    let mut row = 0;
    let mut col = 0;

    for _ in 0..n_links {
        let rowd = rng.gen_range(0..10) as i32 - 3;
        let cold = rng.gen_range(0..10) as i32 - 5;

        row = (row as i32 + rowd).max(0) as usize;
        col = (col as i32 + cold).max(0) as usize;

        k2.add(row, col).unwrap();
    }
}

fn main() {
    println!("Testing LRU with 1000 incremental links...\n");

    // Test without LRU
    let start = std::time::Instant::now();
    let mut k2_no_lru = K2Tree::new_with_config(
        SliceArray::new(),
        SliceArray::new(),
        SIXTEEN_FOUR_CONFIG,
    );
    populate_incremental_tree(1000, &mut k2_no_lru);
    let stats_no_lru = k2_no_lru.stats();
    let duration_no_lru = start.elapsed();

    println!("Without LRU:");
    println!("  Time: {:?}", duration_no_lru);
    println!("  Stats: {:?}\n", stats_no_lru);

    // Test with LRU
    let start = std::time::Instant::now();
    let mut k2_lru = K2Tree::new_with_config(
        LruArray::new(SliceArray::new(), 64),
        LruArray::new(SliceArray::new(), 64),
        SIXTEEN_FOUR_CONFIG,
    );
    populate_incremental_tree(1000, &mut k2_lru);
    let stats_lru = k2_lru.stats();
    let duration_lru = start.elapsed();

    println!("With LRU (cache size 64):");
    println!("  Time: {:?}", duration_lru);
    println!("  Stats: {:?}", stats_lru);
    println!("  Slowdown: {:.2}x", duration_lru.as_secs_f64() / duration_no_lru.as_secs_f64());
}
