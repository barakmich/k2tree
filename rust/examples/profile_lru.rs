use k2tree::{BitArray, K2Tree, LruArray, SliceArray, SIXTEEN_FOUR_CONFIG};
use rand::Rng;
use std::time::Instant;

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
    const N: usize = 10_000;
    const ITERATIONS: usize = 100;

    println!("Profiling LRU vs SliceArray with {} links, {} iterations\n", N, ITERATIONS);

    // Warm up
    {
        let mut k2 = K2Tree::new_with_config(
            SliceArray::new(),
            SliceArray::new(),
            SIXTEEN_FOUR_CONFIG,
        );
        populate_incremental_tree(N, &mut k2);
    }

    // Benchmark SliceArray
    let start = Instant::now();
    for _ in 0..ITERATIONS {
        let mut k2 = K2Tree::new_with_config(
            SliceArray::new(),
            SliceArray::new(),
            SIXTEEN_FOUR_CONFIG,
        );
        populate_incremental_tree(N, &mut k2);
    }
    let slice_time = start.elapsed();
    let slice_avg = slice_time.as_micros() / ITERATIONS as u128;

    println!("SliceArray:");
    println!("  Total: {:?}", slice_time);
    println!("  Average: {} µs\n", slice_avg);

    // Benchmark LRU with cache size 64
    let start = Instant::now();
    for _ in 0..ITERATIONS {
        let mut k2 = K2Tree::new_with_config(
            LruArray::new(SliceArray::new(), 64),
            LruArray::new(SliceArray::new(), 64),
            SIXTEEN_FOUR_CONFIG,
        );
        populate_incremental_tree(N, &mut k2);
    }
    let lru_time = start.elapsed();
    let lru_avg = lru_time.as_micros() / ITERATIONS as u128;

    println!("LRU (cache size 64):");
    println!("  Total: {:?}", lru_time);
    println!("  Average: {} µs", lru_avg);
    println!("  Overhead: {:.2}x slower\n", lru_avg as f64 / slice_avg as f64);

    // Try with different cache sizes
    println!("Testing different cache sizes:");
    for &cache_size in &[8, 16, 32, 64, 128] {
        let start = Instant::now();
        for _ in 0..ITERATIONS {
            let mut k2 = K2Tree::new_with_config(
                LruArray::new(SliceArray::new(), cache_size),
                LruArray::new(SliceArray::new(), cache_size),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(N, &mut k2);
        }
        let time = start.elapsed();
        let avg = time.as_micros() / ITERATIONS as u128;
        println!("  Cache size {}: {} µs ({:.2}x)", cache_size, avg, avg as f64 / slice_avg as f64);
    }

    // Try with different cache distances
    println!("\nTesting different cache distances (cache size 64):");
    for &cache_distance in &[128, 256, 512, 1024, 2048, 4096] {
        let start = Instant::now();
        for _ in 0..ITERATIONS {
            let mut k2 = K2Tree::new_with_config(
                LruArray::with_cache_distance(SliceArray::new(), 64, cache_distance),
                LruArray::with_cache_distance(SliceArray::new(), 64, cache_distance),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(N, &mut k2);
        }
        let time = start.elapsed();
        let avg = time.as_micros() / ITERATIONS as u128;
        println!("  Distance {}: {} µs ({:.2}x)", cache_distance, avg, avg as f64 / slice_avg as f64);
    }

    // Test cache effectiveness
    println!("\nCache effectiveness test:");
    let mut k2_lru = K2Tree::new_with_config(
        LruArray::new(SliceArray::new(), 64),
        LruArray::new(SliceArray::new(), 64),
        SIXTEEN_FOUR_CONFIG,
    );
    populate_incremental_tree(N, &mut k2_lru);

    // Get cache stats (this is a hack - we need to expose the underlying arrays)
    println!("  (Cache stats not directly accessible from K2Tree)");
}
