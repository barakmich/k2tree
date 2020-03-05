package bytearray

import "testing"

func TestRebalanceSpillover(t *testing.T) {
	a := &SpilloverArray{
		bytes:      []byte{0x01, 0x02},
		levelOff:   []int{0},
		levelStart: []int{0},
		levelCount: []int{2},
		length:     2,
		pagesize:   2,
		highwater:  1,
		low:        1,
		multiplier: true,
	}
	a.rebalance()
	a.checkInvariants()

	t.Logf("%#v\n %#v\n", a.bytes, a)

	a = NewSpillover(2, 0.5, 0.5, true)
	a.Insert(0, []byte{0x02})
	t.Logf("%#v\n %#v\n", a.bytes, a)
	a.Insert(0, []byte{0x01})
	a.checkInvariants()
	t.Logf("%#v\n %#v\n", a.bytes, a)
}

func TestCompareSliceSpillover(t *testing.T) {
	tv := insertTestVector()
	vec_a := NewSlice()
	vec_b := NewSpillover(512, 0.75, 0.5, true)
	for i, x := range tv {
		b := byte(i)
		vec_a.Insert(x, []byte{b, b})
		vec_b.Insert(x, []byte{b, b})
		if vec_a.Len() != vec_b.Len() {
			t.Fatalf("Different Lengths after %d: %d %d", i, vec_a.Len(), vec_b.Len())
		}
	}
	vec_b.checkInvariants()

	for i := 0; i < vec_a.Len(); i++ {

		if vec_a.Get(i) != vec_b.Get(i) {
			t.Logf("Spillover Stats: %s", vec_b.stats())
			t.Fatalf("Mismatched byte at %d: ex %v, got %v", i, vec_a.Get(i), vec_b.Get(i))
		}
	}
}
