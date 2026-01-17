/// Inserts the last top four bits (0xF0) of `input` at the first nibble
/// in position 0 in `dest`, and returns a byte with the last four bits (0x0F)
/// of the last byte in `dest` shifted up.
///
pub fn insert_four_bits(dest: &mut [u8], input: u8) -> u8 {
    let mut input = input & 0xF0;
    for i in 0..dest.len() {
        let b = dest[i];
        dest[i] = (b >> 4) | input;
        input = b << 4;
    }
    input
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_insert_four_bits() {
        let mut dest = vec![0xAB, 0xCD];
        let out = insert_four_bits(&mut dest, 0x10);
        assert_eq!(dest, vec![0x1A, 0xBC]);
        assert_eq!(out, 0xD0);
    }

    #[test]
    fn test_insert_four_bits_empty() {
        let mut dest = vec![];
        let out = insert_four_bits(&mut dest, 0xF0);
        assert_eq!(out, 0xF0);
    }

    #[test]
    fn test_insert_four_bits_single() {
        let mut dest = vec![0xFF];
        let out = insert_four_bits(&mut dest, 0xA0);
        assert_eq!(dest, vec![0xAF]);
        assert_eq!(out, 0xF0);
    }
}
