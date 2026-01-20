package k2tree

import "testing"

func TestQuartileCount(t *testing.T) {
	s := newQuartileIndex(&sliceArray{})
	s.Insert(24, 0)
	s.Set(3, true)
	s.Insert(8, 0)
	checkInvariants(t, s)
}

func TestQuartileBugInitialization(t *testing.T) {
	// Create a 16-bit array with bits set across multiple quartiles
	arr := &sliceArray{}
	arr.Insert(16, 0)

	// Set bits in the first quartile (positions 0-3)
	arr.Set(0, true)
	arr.Set(1, true)
	// count[0-4) = 2

	// Set bits in the second quartile (positions 4-7)
	arr.Set(5, true)
	// count[0-8) = 3

	// Set bits in the third quartile (positions 8-11)
	arr.Set(9, true)
	// count[0-12) = 4

	// Set bits in the fourth quartile (positions 12-15)
	arr.Set(14, true)
	// count[0-16) = 5

	t.Logf("Array has %d total set bits", arr.Total())

	// Create the quartile index - this should set counts properly
	q := newQuartileIndex(arr)

	t.Logf("Offsets: %v", q.offsets)
	t.Logf("Counts: %v", q.counts)

	// Check the invariants immediately after construction
	// This will expose the bug in newQuartileIndex
	checkInvariants(t, q)
}

func checkInvariants(t *testing.T, s *quartileIndex) {
	for i, x := range s.offsets {
		expected := s.bits.Count(0, x)
		if expected != s.counts[i] {
			t.Errorf("Count invariant failed: quartile index %d, count %d, expected %d", i, s.counts[i], expected)
		}
	}
}
