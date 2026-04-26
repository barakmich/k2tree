use criterion::{black_box, criterion_group, criterion_main, Criterion};
use k2tree::insert_four_bits;

const INSERT_FOUR_SIZES: &[usize] = &[8, 15, 16, 31, 32, 63, 64, 256, 1024, 4096, 65536];

fn benchmark_insert_four(c: &mut Criterion) {
    let mut group = c.benchmark_group("insert_four");

    for &size in INSERT_FOUR_SIZES {
        let mut buf = vec![0u8; size];
        // Fill with deterministic random-ish data
        for (i, b) in buf.iter_mut().enumerate() {
            *b = ((i * 7 + 13) & 0xFF) as u8;
        }
        let mut inbyte = 0xA0u8;

        group.bench_function(format!("size={}", size), |b| {
            b.iter(|| {
                inbyte = insert_four_bits(black_box(&mut buf), inbyte);
            });
        });
    }

    group.finish();
}

criterion_group!(benches, benchmark_insert_four);
criterion_main!(benches);
