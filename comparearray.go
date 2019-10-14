package k2tree

import "fmt"

type compareArray struct {
	baseline bitarray
	test     bitarray
}

var _ bitarray = (*compareArray)(nil)

func newCompareArray(baseline, test bitarray) *compareArray {
	return &compareArray{
		baseline: baseline,
		test:     test,
	}
}

// Len returns the number of bits in the bitarray.
func (c *compareArray) Len() int {
	a := c.baseline.Len()
	b := c.test.Len()
	assert(a == b, fmt.Sprintf("Len diverged: base: %d test: %d", a, b))
	return a
}

// Set sets the bit at an index `at` to the value `val`.
func (c *compareArray) Set(at int, val bool) {
	c.baseline.Set(at, val)
	c.test.Set(at, val)
}

// Get returns the value stored at `at`.
func (c *compareArray) Get(at int) bool {
	a := c.baseline.Get(at)
	b := c.test.Get(at)
	assert(a == b, fmt.Sprintf("Get diverged on %d: base: %v test: %v", at, a, b))
	return a
}

// Count returns the number of set bits in the interval [from, to).
func (c *compareArray) Count(from int, to int) int {
	a := c.baseline.Count(from, to)
	b := c.test.Count(from, to)
	assert(a == b, fmt.Sprintf("Count diverged on %d, %d: base: %v test: %v", from, to, a, b))
	return a
}

// Total returns the total number of set bits.
func (c *compareArray) Total() int {
	a := c.baseline.Total()
	b := c.test.Total()
	assert(a == b, fmt.Sprintf("Total diverged: base: %v test: %v", a, b))
	return a
}

// Bytes returns the bitarray as a byte array
func (c *compareArray) Bytes() []byte {
	return c.baseline.Bytes()
}

// Insert extends the bitarray by `n` bits. The bits are zeroed
// and start at index `at`. Example:
// Initial string: 11101
// Insert(3, 2)
// Resulting string: 11000101
func (c *compareArray) Insert(n int, at int) error {
	err := c.baseline.Insert(n, at)
	if err != nil {
		return err
	}
	err = c.test.Insert(n, at)
	if err != nil {
		assert(false, "Got an error from test")
	}
	a := c.baseline.Len()
	b := c.test.Len()
	assert(a == b, fmt.Sprintf("Len diverged after Insert: base: %d test: %d", a, b))
	return nil
}

func (c *compareArray) debug() string {
	return fmt.Sprintf("CompareArray\n%s\n%s", c.baseline.debug(), c.test.debug())
}
