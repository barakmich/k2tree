use k2tree::{BitArray, SliceArray};

fn main() {
    let mut arr = SliceArray::new();
    arr.insert(16, 0).unwrap();
    arr.set(0, true);
    arr.set(7, true);
    arr.set(8, true);
    arr.set(11, true);

    println!("Before insert:");
    println!("  get(0) = {}", arr.get(0));
    println!("  get(7) = {}", arr.get(7));
    println!("  get(8) = {}", arr.get(8));
    println!("  get(11) = {}", arr.get(11));
    println!("  count(0, 16) = {}", arr.count(0, 16));

    arr.insert(4, 8).unwrap();

    println!("\nAfter insert(4, 8):");
    println!("  len = {}", arr.len());
    println!("  get(0) = {}", arr.get(0));
    println!("  get(7) = {}", arr.get(7));
    println!("  get(8) = {}", arr.get(8));
    println!("  get(11) = {}", arr.get(11));
    println!("  get(12) = {}", arr.get(12));
    println!("  get(15) = {}", arr.get(15));
    println!("  count(0, 20) = {}", arr.count(0, 20));
}
