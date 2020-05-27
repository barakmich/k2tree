package k2tree

// insertFourBitsGo inserts the last top four bits (0xF0) of in
// at the first nibble in position 0 in dest, and returns a byte
// with the last four bits (0x0F) of dest shifted up.
func insertFourBitsGo(dest []byte, in byte) (out byte) {
	in = in & 0xF0
	for i := 0; i < len(dest); i++ {
		b := dest[i]
		dest[i] = b>>4 | in
		in = b << 4
	}
	return in
}
