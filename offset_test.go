package k2tree

import "testing"

func TestFourBOffset(t *testing.T) {
	k2 := &K2Tree{
		tk: fourBitsPerLayer,
		lk: fourBitsPerLayer,
	}

	tt := []struct {
		i          int
		j          int
		expectedL  int
		expectedT1 int
		expectedT2 int
	}{
		{
			i:          0,
			j:          0,
			expectedL:  0,
			expectedT1: 0,
			expectedT2: 0,
		},
		{
			i:          2,
			j:          2,
			expectedL:  0,
			expectedT1: 3,
			expectedT2: 0,
		},
		{
			i:          0,
			j:          6,
			expectedL:  0,
			expectedT1: 1,
			expectedT2: 1,
		},
		{
			i:          4,
			j:          4,
			expectedL:  0,
			expectedT1: 0,
			expectedT2: 3,
		},
		{
			i:          4,
			j:          5,
			expectedL:  1,
			expectedT1: 0,
			expectedT2: 3,
		},
	}

	for _, x := range tt {
		out := k2.offsetL(x.i, x.j)
		if out != x.expectedL {
			t.Errorf("unexpected result %d for L on test %#v\n", out, x)
		}
	}

	for _, x := range tt {
		out := k2.offsetTForLayer(x.i, x.j, 1)
		if out != x.expectedT1 {
			t.Errorf("unexpected result %d for T1 on test %#v\n", out, x)
		}
	}

	for _, x := range tt {
		out := k2.offsetTForLayer(x.i, x.j, 2)
		if out != x.expectedT2 {
			t.Errorf("unexpected result %d for T2 on test %#v\n", out, x)
		}
	}

}
