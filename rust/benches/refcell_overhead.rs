use criterion::{black_box, criterion_group, criterion_main, Criterion};
use std::cell::RefCell;

fn bench_direct_access(c: &mut Criterion) {
    let mut vec = vec![0usize; 64];

    c.bench_function("direct_vec_access", |b| {
        b.iter(|| {
            for i in 0..1000 {
                let idx = black_box(i % 64);
                vec[idx] = vec[idx].wrapping_add(1);
            }
        });
    });
}

fn bench_refcell_access(c: &mut Criterion) {
    let vec = RefCell::new(vec![0usize; 64]);

    c.bench_function("refcell_vec_access", |b| {
        b.iter(|| {
            for i in 0..1000 {
                let idx = black_box(i % 64);
                let mut borrowed = vec.borrow_mut();
                borrowed[idx] = borrowed[idx].wrapping_add(1);
            }
        });
    });
}

fn bench_refcell_single_borrow(c: &mut Criterion) {
    let vec = RefCell::new(vec![0usize; 64]);

    c.bench_function("refcell_vec_single_borrow", |b| {
        b.iter(|| {
            let mut borrowed = vec.borrow_mut();
            for i in 0..1000 {
                let idx = black_box(i % 64);
                borrowed[idx] = borrowed[idx].wrapping_add(1);
            }
        });
    });
}

criterion_group!(
    benches,
    bench_direct_access,
    bench_refcell_access,
    bench_refcell_single_borrow
);
criterion_main!(benches);
