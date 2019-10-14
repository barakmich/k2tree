package k2tree

import (
	"fmt"
	"testing"
)

var curFunc newBitArrayFunc

var testFuncs []func(t *testing.T) = []func(t *testing.T){
	testSmoke,
	testEasyInsert,
	testByteInsert,
	testNibbleInsert,
}

func TestBitarrayTypes(t *testing.T) {
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
				return newPagedSliceArray(10)
			},
			name: "PagedSlice(10)",
		},
		{
			create: func() bitarray {
				return newPagedSliceArray(1000)
			},
			name: "PagedSlice(1000)",
		},
		{
			create: func() bitarray {
				return newQuartileIndex(&sliceArray{})
			},
			name: "QuartileIndex(sliceArray)",
		},
		{
			create: func() bitarray {
				return newInt16Index(&sliceArray{})
			},
			name: "Int16Index(sliceArray)",
		},
	}
	for _, bitarray := range tt {
		curFunc = bitarray.create
		for _, testcase := range testFuncs {
			t.Run(fmt.Sprintf("%s::%s", bitarray.name, GetFunctionName(testcase)), testcase)
		}
	}
}

func testSmoke(t *testing.T) {
	s := curFunc()
	s.Insert(24, 0)
	s.Set(3, true)
	fmt.Println(s.debug())
	if s.Count(0, 2) != 0 {
		t.Error("wrong count")
	}
	if s.Count(2, 3) != 0 {
		t.Error("end inclusive?")
	}
	if s.Count(2, 4) != 1 {
		t.Error("can't count?")
	}
	if !s.Get(3) {
		t.Error("can't retrieve?")
	}
	for x := 0; x < 24; x++ {
		s.Set(x, true)
	}
	fmt.Println(s.debug())
	if s.Count(0, 8) != 8 {
		t.Error("wrong count")
	}
	if s.Count(0, s.Len()) != 24 {
		t.Error("wrong count")
	}
}

func testEasyInsert(t *testing.T) {
	s := curFunc()
	s.Insert(24, 0)
	s.Set(3, true)
	s.Insert(8, 0)
	fmt.Println(s.debug())
	if s.Get(3) {
		t.Error("new 3 should not be set")
	}
	if !s.Get(11) {
		t.Error("new 11 should be set")
	}
	if s.Count(0, 32) != 1 {
		t.Error("count is incorrect -- only one bit was set")
	}
}

func testByteInsert(t *testing.T) {
	s := curFunc()
	s.Insert(24, 0)
	s.Set(11, true)
	s.Set(6, true)
	s.Set(2, true)
	fmt.Println(s.debug())
	s.Insert(8, 4)
	fmt.Println(s.debug())
	if s.Get(11) {
		t.Error("new 11 should not be set")
	}
	if !s.Get(19) {
		t.Error("new 19 should be set")
	}
	if !s.Get(14) {
		t.Error("new 14 should be set")
	}
	if !s.Get(2) {
		t.Error("new 2 should be set")
	}

}

func testNibbleInsert(t *testing.T) {
	s := curFunc()
	s.Insert(24, 0)
	s.Set(11, true)
	s.Set(6, true)
	s.Set(2, true)
	fmt.Println(s.debug())
	s.Insert(4, 4)
	fmt.Println(s.debug())
	if s.Get(11) {
		t.Error("new 11 should not be set")
	}
	if !s.Get(15) {
		t.Error("new 15 should be set")
	}
	if !s.Get(10) {
		t.Error("new 10 should be set")
	}
	if !s.Get(2) {
		t.Error("new 2 should be set")
	}

}

func testNibbleInsertAtZero(t *testing.T) {
	s := curFunc()
	s.Insert(4, 0)
	s.Set(3, true)
	s.Set(0, true)
	s.Insert(4, 0)
	fmt.Println(s.debug())
	if s.Get(0) {
		t.Error("got a wrong 0")
	}
	if s.Get(3) {
		t.Error("got a wrong 3")
	}
	if !s.Get(4) {
		t.Error("got no 4")
	}
	if !s.Get(7) {
		t.Error("got no 7")
	}
}
