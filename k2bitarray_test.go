package k2tree

import "testing"

func TestPopulate50k(t *testing.T) {
	tt := []struct {
		create newBitArrayFunc
		name   string
	}{
		{
			create: func() bitarray {
				return &sliceArray{}
			},
			name: "SliceArray",
		},
		{
			create: func() bitarray {
				return newPagedSliceArray(1000)
			},
			name: "PagedSlice1000",
		},
		{
			create: func() bitarray {
				return newQuartileIndex(&sliceArray{})
			},
			name: "QuartileIndex",
		},
		{
			create: func() bitarray {
				return newInt16Index(&sliceArray{})
			},
			name: "Int16Index",
		},
		{
			create: func() bitarray {
				return newBinaryLRUIndex(&sliceArray{}, 2)
			},
			name: "BinaryLRU2",
		},
	}
	for _, test := range tt {
		t.Log(test.name)
		k2, err := newK2Tree(test.create, Config{
			TreeLayerDef: SixteenBitsPerLayer,
			CellLayerDef: FourBitsPerLayer,
		})
		if err != nil {
			t.Fatal(err)
		}
		populateRandomTree(50000, 50000, k2)
	}
}

func TestRandAdd(t *testing.T) {
	k2, err := newK2Tree(
		func() bitarray {
			return newPagedSliceArray(10)
		},
		Config{
			TreeLayerDef: SixteenBitsPerLayer,
			CellLayerDef: FourBitsPerLayer,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	k2.Add(48081, 27887)
	k2.Add(31847, 34059)
	k2.Add(2081, 41318)
	k2.Add(4425, 22540)
	k2.Add(40456, 3300)
	tmp := k2.Row(2081).ExtractAll()
	if len(tmp) != 1 && tmp[0] != 41318 {
		t.Errorf("Unmatched 2081")
	}
	tmp = k2.Row(40456).ExtractAll()
	if len(tmp) != 1 && tmp[0] != 3300 {
		t.Errorf("Unmatched 40456")
	}

}
