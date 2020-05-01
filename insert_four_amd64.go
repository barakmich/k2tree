// +build amd64,!gccgo,!appengine

package k2tree

// insertFourBits inserts the last four bits (0x0F) of in
// at position 0 in dest, and returns a byte with the last four bits
// (0x0F) of dest shifted up.
func insertFourBits(dest []byte, in byte) byte {
	if len(dest) == 0 {
		return in << 4
	}
	return insertFourBitsAsm(&dest[0], uint64(len(dest)), in)
}

//go:noescape
func insertFourBitsAsm(src *byte, len uint64, in byte) (ret byte)
