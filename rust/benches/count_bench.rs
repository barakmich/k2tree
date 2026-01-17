use criterion::{black_box, criterion_group, criterion_main, Criterion};
use k2tree::{BitArray, BitVecArray, SliceArray};

fn bench_count_small_slice(c: &mut Criterion) {
    let mut arr = SliceArray::new();
    arr.insert(1000, 0).unwrap();
    // Set every 3rd bit
    for i in (0..1000).step_by(3) {
        arr.set(i, true);
    }

    c.bench_function("count_small_slice", |b| {
        b.iter(|| {
            let result = arr.count(black_box(0), black_box(1000));
            black_box(result);
        });
    });
}

fn bench_count_large_slice(c: &mut Criterion) {
    let mut arr = SliceArray::new();
    arr.insert(100_000, 0).unwrap();
    // Set every 3rd bit
    for i in (0..100_000).step_by(3) {
        arr.set(i, true);
    }

    c.bench_function("count_large_slice", |b| {
        b.iter(|| {
            let result = arr.count(black_box(0), black_box(100_000));
            black_box(result);
        });
    });
}

fn bench_count_small_bitvec(c: &mut Criterion) {
    let mut arr = BitVecArray::new();
    arr.insert(1000, 0).unwrap();
    // Set every 3rd bit
    for i in (0..1000).step_by(3) {
        arr.set(i, true);
    }

    c.bench_function("count_small_bitvec", |b| {
        b.iter(|| {
            let result = arr.count(black_box(0), black_box(1000));
            black_box(result);
        });
    });
}

fn bench_count_large_bitvec(c: &mut Criterion) {
    let mut arr = BitVecArray::new();
    arr.insert(100_000, 0).unwrap();
    // Set every 3rd bit
    for i in (0..100_000).step_by(3) {
        arr.set(i, true);
    }

    c.bench_function("count_large_bitvec", |b| {
        b.iter(|| {
            let result = arr.count(black_box(0), black_box(100_000));
            black_box(result);
        });
    });
}

criterion_group!(
    benches,
    bench_count_small_slice,
    bench_count_large_slice,
    bench_count_small_bitvec,
    bench_count_large_bitvec,
);
criterion_main!(benches);
