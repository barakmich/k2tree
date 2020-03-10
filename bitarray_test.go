package k2tree

import (
	"fmt"
	"testing"

	"git.barakmich.com/barak/k2tree/bytearray"
)

var curFunc newBitArrayFunc

type testFunc struct {
	testcase func(t *testing.T)
	name     string
}

type bitArrayType struct {
	create newBitArrayFunc
	name   string
}

var testFuncs []testFunc = []testFunc{
	{testSmoke, "TestSmoke"},
	{testEasyInsert, "TestEasyInsert"},
	{testByteInsert, "TestByteInsert"},
	{testNibbleInsert, "TestNibbleInsert"},
	{testDebug, "TestDebug"},
}

var debugBitArrayTypes []bitArrayType = []bitArrayType{
	{
		create: func() bitarray {
			return newCompareArray(&sliceArray{}, newPagedSliceArray(10))
		},
		name: "CompSlicePaged10",
	},
	{
		create: func() bitarray {
			return newTraceArray(&sliceArray{})
		},
		name: "Trace",
	},
	{
		create: func() bitarray {
			return newDebugArray(&sliceArray{})
		},
		name: "Debug",
	},
}

var testBitArrayTypes []bitArrayType = []bitArrayType{
	{
		create: func() bitarray {
			return &sliceArray{}
		},
		name: "Slice",
	},
	{
		create: func() bitarray {
			return newByteArray(bytearray.NewSlice())
		},
		name: "ByteArray",
	},
	{
		create: func() bitarray {
			return newPagedSliceArray(1000)
		},
		name: "Paged1k",
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
		name: "Int16",
	},
	{
		create: func() bitarray {
			return newByteArray(bytearray.NewInt16Index(bytearray.NewSlice()))
		},
		name: "BAInt16",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(&sliceArray{}, 2)
		},
		name: "LRU2",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(&sliceArray{}, 128)
		},
		name: "LRU128",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newPagedSliceArray(50000), 128)
		},
		name: "LRU128Paged50k",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newPagedSliceArray(100000), 128)
		},
		name: "LRU128Paged100k",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newByteArray(bytearray.NewSpillover(16*1024, 0.9, 0.4, false)), 128)
		},
		name: "LRU128Spill16k1x",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newByteArray(bytearray.NewSpillover(16*1024, 0.9, 0.4, true)), 128)
		},
		name: "LRU128Spill16k2x",
	},
}

func TestBitarrayTypes(t *testing.T) {
	for _, bitarray := range testBitArrayTypes {
		curFunc = bitarray.create
		for _, testcase := range testFuncs {
			t.Run(fmt.Sprintf("%s%s", testcase.name, bitarray.name), testcase.testcase)
		}
	}
	for _, bitarray := range debugBitArrayTypes {
		curFunc = bitarray.create
		for _, testcase := range testFuncs {
			t.Run(fmt.Sprintf("%s%s", testcase.name, bitarray.name), testcase.testcase)
		}
	}
}

func testSmoke(t *testing.T) {
	s := curFunc()
	s.Insert(24, 0)
	s.Set(3, true)
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
	if s.Count(0, 8) != 8 {
		t.Error("wrong count")
	}
	if s.Count(0, s.Len()) != 24 {
		t.Error("wrong count")
	}
	if s.Total() != 24 {
		t.Error("wrong total")
	}
}

func testEasyInsert(t *testing.T) {
	s := curFunc()
	s.Insert(24, 0)
	s.Set(3, true)
	s.Insert(8, 0)
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
	s.Insert(8, 4)
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
	s.Insert(4, 4)
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

func testDebug(t *testing.T) {
	s := curFunc()
	s.debug()
}
