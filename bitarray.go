package k2tree

type bitarray interface {
	// Len returns the number of bits in the bitarray.
	Len() int
	// Set sets the bit at an index `at` to the value `val`.
	Set(at int, val bool)
	// Get returns the value stored at `at`.
	Get(at int) bool
	// Count returns the number of set bits in the interval [from, to).
	Count(from, to int) int
	// Total returns the total number of set bits.
	Total() int
	// Insert extends the bitarray by `n` bits. The bits are zeroed
	// and start at index `at`. Example:
	// Initial string: 11101
	// Insert(3, 2)
	// Resulting string: 11000101
	Insert(n int, at int) error
	debug() string
}

type newBitArrayFunc func() bitarray
