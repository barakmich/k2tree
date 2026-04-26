/// Inserts the last top four bits (0xF0) of `input` at the first nibble
/// in position 0 in `dest`, and returns a byte with the last four bits (0x0F)
/// of the last byte in `dest` shifted up.
///
/// On aarch64 this uses NEON intrinsics to process 16 bytes per iteration.
/// On other targets it falls back to a scalar loop.
#[cfg(target_arch = "aarch64")]
pub fn insert_four_bits(dest: &mut [u8], input: u8) -> u8 {
    unsafe { insert_four_bits_neon(dest, input) }
}

#[cfg(not(target_arch = "aarch64"))]
pub fn insert_four_bits(dest: &mut [u8], input: u8) -> u8 {
    insert_four_bits_scalar(dest, input)
}

/// Scalar fallback for non-NEON targets.
#[allow(dead_code)]
fn insert_four_bits_scalar(dest: &mut [u8], input: u8) -> u8 {
    let mut input = input & 0xF0;
    for i in 0..dest.len() {
        let b = dest[i];
        dest[i] = (b >> 4) | input;
        input = b << 4;
    }
    input
}

/// NEON-accelerated insert_four_bits for aarch64.
///
/// Processes 16 bytes per iteration:
///   result[i] = (src[i] >> 4) | (src[i-1] << 4)    for i in 1..15
///   result[0] = (src[0] >> 4) | carry
///   new carry = (src[15] << 4) & 0xF0
///
/// Uses a vector "carry register" with the carry byte stored at index 15,
/// so VEXT can slide it in without a round-trip to scalar.
#[cfg(target_arch = "aarch64")]
#[target_feature(enable = "neon")]
unsafe fn insert_four_bits_neon(dest: &mut [u8], input: u8) -> u8 {
    use std::arch::aarch64::*;

    let mut ptr = dest.as_mut_ptr();
    let mut remaining = dest.len();
    let input = input & 0xF0;

    // V3 = zero with carry placed in byte 15.
    let zero = vdupq_n_u8(0);
    let mut carry = vsetq_lane_u8(input, zero, 15);

    // Main loop: 16 bytes per iteration.
    while remaining >= 16 {
        let v0 = vld1q_u8(ptr);
        let shr4 = vshrq_n_u8(v0, 4); // V1[i] = V0[i] >> 4
        let shl4 = vshlq_n_u8(v0, 4); // V0[i] = V0[i] << 4
                                      // V2 = [carry, shl4[0], shl4[1], ..., shl4[14]]
        let v2 = vextq_u8(shl4, carry, 15);
        let result = vorrq_u8(shr4, v2);
        // Stash shl4[15] as the next carry in V3[15].
        carry = vsetq_lane_u8(vgetq_lane_u8(shl4, 15), carry, 15);
        vst1q_u8(ptr, result);

        ptr = ptr.add(16);
        remaining -= 16;
    }

    // Extract carry back to scalar for the tail.
    let mut carry_byte = vgetq_lane_u8(carry, 15);

    // Tail: scalar loop for remaining bytes.
    let tail = std::slice::from_raw_parts_mut(ptr, remaining);
    for b in tail.iter_mut() {
        let old = *b;
        *b = (old >> 4) | carry_byte;
        carry_byte = old << 4;
    }

    carry_byte
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
