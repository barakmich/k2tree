package k2tree

import "testing"

func TestQuartileCount(t *testing.T) {
	s := newQuartileIndex(&sliceArray{})
	s.Insert(24, 0)
	s.Set(3, true)
	s.Insert(8, 0)
	checkInvariants(t, s)
}

func checkInvariants(t *testing.T, s *quartileIndex) {
	for i, x := range s.offsets {
		expected := s.bits.Count(0, x)
		if expected != s.counts[i] {
			t.Errorf("Count invariant failed: quartile index %d, count %d, expected %d", i, s.counts[i], expected)
		}
	}
}
