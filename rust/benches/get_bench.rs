use criterion::{black_box, criterion_group, criterion_main, Criterion};
use k2tree::{BitArray, BitVecArray, SliceArray};

fn bench_get_sequential_slice(c: &mut Criterion) {
    let mut arr = SliceArray::new();
    arr.insert(10000, 0).unwrap();
    for i in (0..10000).step_by(3) {
        arr.set(i, true);
    }

    c.bench_function("get_sequential_slice", |b| {
        b.iter(|| {
            let mut sum = 0;
            for i in 0..10000 {
                if arr.get(black_box(i)) {
                    sum += 1;
                }
            }
            black_box(sum);
        });
    });
}

fn bench_get_sequential_bitvec(c: &mut Criterion) {
    let mut arr = BitVecArray::new();
    arr.insert(10000, 0).unwrap();
    for i in (0..10000).step_by(3) {
        arr.set(i, true);
    }

    c.bench_function("get_sequential_bitvec", |b| {
        b.iter(|| {
            let mut sum = 0;
            for i in 0..10000 {
                if arr.get(black_box(i)) {
                    sum += 1;
                }
            }
            black_box(sum);
        });
    });
}

fn bench_get_random_slice(c: &mut Criterion) {
    let mut arr = SliceArray::new();
    arr.insert(10000, 0).unwrap();
    for i in (0..10000).step_by(3) {
        arr.set(i, true);
    }

    // Pre-generate random indices
    let indices: Vec<usize> = (0..10000).map(|i| (i * 7919) % 10000).collect();

    c.bench_function("get_random_slice", |b| {
        b.iter(|| {
            let mut sum = 0;
            for &i in &indices {
                if arr.get(black_box(i)) {
                    sum += 1;
                }
            }
            black_box(sum);
        });
    });
}

fn bench_get_random_bitvec(c: &mut Criterion) {
    let mut arr = BitVecArray::new();
    arr.insert(10000, 0).unwrap();
    for i in (0..10000).step_by(3) {
        arr.set(i, true);
    }

    // Pre-generate random indices
    let indices: Vec<usize> = (0..10000).map(|i| (i * 7919) % 10000).collect();

    c.bench_function("get_random_bitvec", |b| {
        b.iter(|| {
            let mut sum = 0;
            for &i in &indices {
                if arr.get(black_box(i)) {
                    sum += 1;
                }
            }
            black_box(sum);
        });
    });
}

criterion_group!(
    benches,
    bench_get_sequential_slice,
    bench_get_sequential_bitvec,
    bench_get_random_slice,
    bench_get_random_bitvec,
);
criterion_main!(benches);
