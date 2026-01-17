use criterion::{black_box, criterion_group, criterion_main, Criterion};
use k2tree::{BitArray, BitVecArray, K2Tree, SliceArray, SIXTEEN_FOUR_CONFIG, SIXTEEN_SIXTEEN_CONFIG};
use rand::Rng;

fn simple_load<T: BitArray>(k: &mut K2Tree<T>) {
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

fn populate_random_tree<T: BitArray>(n_links: usize, max_id: usize, k2: &mut K2Tree<T>) {
    let mut rng = rand::thread_rng();
    for _ in 0..n_links {
        let row = rng.gen_range(0..max_id);
        let col = rng.gen_range(0..max_id);
        k2.add(row, col).unwrap();
    }
}

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

// SliceArray benchmarks

fn bench_extract_20_slice(c: &mut Criterion) {
    let mut k2 = K2Tree::new(SliceArray::new(), SliceArray::new());
    simple_load(&mut k2);

    c.bench_function("extract_20_slice", |b| {
        b.iter(|| {
            let mut it = k2.from(black_box(20));
            let out = it.extract_all();
            black_box(out);
        });
    });
}

fn bench_rand_pop_1k_slice(c: &mut Criterion) {
    c.bench_function("rand_pop_1k_slice", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                SliceArray::new(),
                SliceArray::new(),
                SIXTEEN_SIXTEEN_CONFIG,
            );
            populate_random_tree(black_box(1000), black_box(2000), &mut k2);
            black_box(k2.stats());
        });
    });
}

fn bench_inc_pop_1k_slice(c: &mut Criterion) {
    c.bench_function("inc_pop_1k_slice", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                SliceArray::new(),
                SliceArray::new(),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(black_box(1000), &mut k2);
            black_box(k2.stats());
        });
    });
}

fn bench_rand_pop_10k_slice(c: &mut Criterion) {
    c.bench_function("rand_pop_10k_slice", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                SliceArray::new(),
                SliceArray::new(),
                SIXTEEN_SIXTEEN_CONFIG,
            );
            populate_random_tree(black_box(10000), black_box(20000), &mut k2);
            let stats = k2.stats();
            black_box(stats);
        });
    });
}

fn bench_inc_pop_10k_slice(c: &mut Criterion) {
    c.bench_function("inc_pop_10k_slice", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                SliceArray::new(),
                SliceArray::new(),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(black_box(10000), &mut k2);
            let stats = k2.stats();
            black_box(stats);
        });
    });
}

// BitVecArray benchmarks

fn bench_extract_20_bitvec(c: &mut Criterion) {
    let mut k2 = K2Tree::new(BitVecArray::new(), BitVecArray::new());
    simple_load(&mut k2);

    c.bench_function("extract_20_bitvec", |b| {
        b.iter(|| {
            let mut it = k2.from(black_box(20));
            let out = it.extract_all();
            black_box(out);
        });
    });
}

fn bench_rand_pop_1k_bitvec(c: &mut Criterion) {
    c.bench_function("rand_pop_1k_bitvec", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                BitVecArray::new(),
                BitVecArray::new(),
                SIXTEEN_SIXTEEN_CONFIG,
            );
            populate_random_tree(black_box(1000), black_box(2000), &mut k2);
            black_box(k2.stats());
        });
    });
}

fn bench_inc_pop_1k_bitvec(c: &mut Criterion) {
    c.bench_function("inc_pop_1k_bitvec", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                BitVecArray::new(),
                BitVecArray::new(),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(black_box(1000), &mut k2);
            black_box(k2.stats());
        });
    });
}

fn bench_rand_pop_10k_bitvec(c: &mut Criterion) {
    c.bench_function("rand_pop_10k_bitvec", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                BitVecArray::new(),
                BitVecArray::new(),
                SIXTEEN_SIXTEEN_CONFIG,
            );
            populate_random_tree(black_box(10000), black_box(20000), &mut k2);
            let stats = k2.stats();
            black_box(stats);
        });
    });
}

fn bench_inc_pop_10k_bitvec(c: &mut Criterion) {
    c.bench_function("inc_pop_10k_bitvec", |b| {
        b.iter(|| {
            let mut k2 = K2Tree::new_with_config(
                BitVecArray::new(),
                BitVecArray::new(),
                SIXTEEN_FOUR_CONFIG,
            );
            populate_incremental_tree(black_box(10000), &mut k2);
            let stats = k2.stats();
            black_box(stats);
        });
    });
}

criterion_group!(
    benches,
    bench_extract_20_slice,
    bench_rand_pop_1k_slice,
    bench_inc_pop_1k_slice,
    bench_rand_pop_10k_slice,
    bench_inc_pop_10k_slice,
    bench_extract_20_bitvec,
    bench_rand_pop_1k_bitvec,
    bench_inc_pop_1k_bitvec,
    bench_rand_pop_10k_bitvec,
    bench_inc_pop_10k_bitvec,
);
criterion_main!(benches);
