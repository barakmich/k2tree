package k2tree

// insertFourBits inserts the last four bits (0x0F) of in
// at position 0 in dest, and returns a byte with the last four bits
// (0x0F) of dest shifted up.
func insertFourBits(dest []byte, in byte) (out byte) {
	in = in << 4
	for i := 0; i < len(dest); i++ {
		b := dest[i]
		dest[i] = b>>4 | in
		in = b << 4
	}
	return in
}
