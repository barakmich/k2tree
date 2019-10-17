package k2tree

import (
	"fmt"
	"sort"
	"testing"
)

func simpleLoad(k *K2Tree) {
	k.Add(20, 41)
	k.Add(14, 20)
	k.Add(20, 2)
	k.Add(20, 1)
	k.Add(1, 14)
	k.Add(20, 14)
	k.Add(20, 30)
	k.Add(30, 30)
	k.Add(20, 17)
	k.Add(41, 17)
	k.Add(41, 1)
	k.Add(41, 30)
}

func TestRowIterator(t *testing.T) {
	tt := []struct {
		loadtree func(*K2Tree)
		row      int
		expected []int
	}{
		//{
		//loadtree: simpleLoad,
		//row:      20,
		//expected: []int{1, 2, 14, 17, 30, 41},
		//},
		{
			loadtree: simpleLoad,
			row:      41,
			expected: []int{1, 17, 30},
		},
	}

	for i, test := range tt {
		k2, err := newK2Tree(func() bitarray { return &sliceArray{} }, DefaultConfig)
		if err != nil {
			t.Fatal(err)
		}
		test.loadtree(k2)
		fmt.Println(k2.Stats())
		it := newRowIterator(k2, test.row)
		var out []int
		for it.Next() {
			out = append(out, it.Value())
		}
		sort.Ints(out)
		if len(test.expected) != len(out) {
			t.Errorf("instance %d mismatch in length: out: %v expected %v", i, out, test.expected)
		}
		for i := range test.expected {
			if test.expected[i] != out[i] {
				t.Errorf("instance %d mismatch: out: %v expected: %v", i, out, test.expected)
			}
		}

	}
}
