// +build !amd64

package k2tree

// insertFourBits inserts the last four bits (0x0F) of in
// at position 0 in dest, and returns a byte with the last four bits
// (0x0F) of dest shifted up.
func insertFourBits(dest []byte, in byte) (out byte) {
	return insertFourBitsGo(dest, in)
}
