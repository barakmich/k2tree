package bytearray

import "testing"

func TestRebalanceSpillover(t *testing.T) {
	a := &SpilloverArray{
		bytes:      []byte{0x01, 0x02},
		levelOff:   []int{0},
		levelCum:   []int{2},
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
	a.Insert(0, []byte{0x01})
	a.Insert(0, []byte{0x01})
	a.Insert(0, []byte{0x01})
	a.checkInvariants()
	t.Logf("%#v\n %#v\n", a.bytes, a)
}

func TestCompareSliceSpillover(t *testing.T) {
	vec := NewSpillover(512, 0.75, 0.5, true)
	testCompareBaseline(t, vec)
}
