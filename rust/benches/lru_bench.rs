use criterion::{black_box, criterion_group, criterion_main, Criterion};
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

// Cache size sweep benchmarks
// Testing different LRU cache sizes to find optimal value

fn bench_lru_size_16(c: &mut Criterion) {
    c.bench_function("lru_size_16", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                LruArray::new(SliceArray::new(), 16),
                LruArray::new(SliceArray::new(), 16),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(black_box(10_000), &mut k2);
            black_box(k2.stats());
        });
    });
}

fn bench_lru_size_32(c: &mut Criterion) {
    c.bench_function("lru_size_32", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                LruArray::new(SliceArray::new(), 32),
                LruArray::new(SliceArray::new(), 32),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(black_box(10_000), &mut k2);
            black_box(k2.stats());
        });
    });
}

fn bench_lru_size_64(c: &mut Criterion) {
    c.bench_function("lru_size_64", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                LruArray::new(SliceArray::new(), 64),
                LruArray::new(SliceArray::new(), 64),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(black_box(10_000), &mut k2);
            black_box(k2.stats());
        });
    });
}

fn bench_lru_size_128(c: &mut Criterion) {
    c.bench_function("lru_size_128", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                LruArray::new(SliceArray::new(), 128),
                LruArray::new(SliceArray::new(), 128),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(black_box(10_000), &mut k2);
            black_box(k2.stats());
        });
    });
}

fn bench_lru_size_256(c: &mut Criterion) {
    c.bench_function("lru_size_256", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                LruArray::new(SliceArray::new(), 256),
                LruArray::new(SliceArray::new(), 256),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(black_box(10_000), &mut k2);
            black_box(k2.stats());
        });
    });
}

fn bench_lru_size_512(c: &mut Criterion) {
    c.bench_function("lru_size_512", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                LruArray::new(SliceArray::new(), 512),
                LruArray::new(SliceArray::new(), 512),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(black_box(10_000), &mut k2);
            black_box(k2.stats());
        });
    });
}

// Cache distance sweep benchmarks
// Testing different cache distance thresholds

fn bench_lru_dist_256(c: &mut Criterion) {
    c.bench_function("lru_dist_256", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                LruArray::with_cache_distance(SliceArray::new(), 64, 256),
                LruArray::with_cache_distance(SliceArray::new(), 64, 256),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(black_box(10_000), &mut k2);
            black_box(k2.stats());
        });
    });
}

fn bench_lru_dist_512(c: &mut Criterion) {
    c.bench_function("lru_dist_512", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                LruArray::with_cache_distance(SliceArray::new(), 64, 512),
                LruArray::with_cache_distance(SliceArray::new(), 64, 512),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(black_box(10_000), &mut k2);
            black_box(k2.stats());
        });
    });
}

fn bench_lru_dist_1024(c: &mut Criterion) {
    c.bench_function("lru_dist_1024", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                LruArray::with_cache_distance(SliceArray::new(), 64, 1024),
                LruArray::with_cache_distance(SliceArray::new(), 64, 1024),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(black_box(10_000), &mut k2);
            black_box(k2.stats());
        });
    });
}

fn bench_lru_dist_2048(c: &mut Criterion) {
    c.bench_function("lru_dist_2048", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                LruArray::with_cache_distance(SliceArray::new(), 64, 2048),
                LruArray::with_cache_distance(SliceArray::new(), 64, 2048),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(black_box(10_000), &mut k2);
            black_box(k2.stats());
        });
    });
}

fn bench_lru_dist_4096(c: &mut Criterion) {
    c.bench_function("lru_dist_4096", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                LruArray::with_cache_distance(SliceArray::new(), 64, 4096),
                LruArray::with_cache_distance(SliceArray::new(), 64, 4096),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(black_box(10_000), &mut k2);
            black_box(k2.stats());
        });
    });
}

// Baseline comparison
fn bench_no_lru(c: &mut Criterion) {
    c.bench_function("no_lru_baseline", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                SliceArray::new(),
                SliceArray::new(),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(black_box(10_000), &mut k2);
            black_box(k2.stats());
        });
    });
}

criterion_group!(
    lru_cache_size,
    bench_no_lru,
    bench_lru_size_16,
    bench_lru_size_32,
    bench_lru_size_64,
    bench_lru_size_128,
    bench_lru_size_256,
    bench_lru_size_512,
);

criterion_group!(
    lru_cache_distance,
    bench_lru_dist_256,
    bench_lru_dist_512,
    bench_lru_dist_1024,
    bench_lru_dist_2048,
    bench_lru_dist_4096,
);

criterion_main!(lru_cache_size, lru_cache_distance);
