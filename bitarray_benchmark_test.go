package k2tree

import (
	"fmt"
	"testing"

	"git.barakmich.com/barak/k2tree/bytearray"
)

func testPopulateRand(t testing.TB, ba newBitArrayFunc, n int, compare bool) *K2Tree {
	createF := ba
	if compare {
		createF = func() bitarray {
			x := ba()
			return newCompareArray(&sliceArray{}, x)
		}
	}
	k2, err := newK2Tree(
		createF,
		Config{
			TreeLayerDef: SixteenBitsPerLayer,
			CellLayerDef: SixteenBitsPerLayer,
		})
	if err != nil {
		t.Fatal(err)
	}
	populateRandomTree(n, n*2, k2, false)
	return k2
}

func testPopulateIncremental(t testing.TB, ba newBitArrayFunc, n int, compare bool) *K2Tree {
	createF := ba
	if compare {
		createF = func() bitarray {
			x := ba()
			return newCompareArray(&sliceArray{}, x)
		}
	}
	k2, err := newK2Tree(
		createF,
		Config{
			TreeLayerDef: SixteenBitsPerLayer,
			CellLayerDef: SixteenBitsPerLayer,
		})
	if err != nil {
		t.Fatal(err)
	}
	populateIncrementalTree(n, k2, false)
	return k2
}

func TestRandPop50k(t *testing.T) {
	for _, bitarray := range testBitArrayTypes {
		t.Run(fmt.Sprintf(bitarray.name), func(t *testing.T) {
			k2 := testPopulateRand(t, bitarray.create, 50000, true)
			t.Logf("%f bpl", k2.Stats().BitsPerLink)
		})
	}
}

func TestIncPop50k(t *testing.T) {
	for _, bitarray := range testBitArrayTypes {
		t.Run(fmt.Sprintf(bitarray.name), func(t *testing.T) {
			k2 := testPopulateIncremental(t, bitarray.create, 50000, true)
			t.Logf("%f bpl", k2.Stats().BitsPerLink)
		})
	}
}

func BenchmarkRandPop50k(b *testing.B) {
	for _, bitarray := range testBitArrayTypes {
		b.Run(fmt.Sprintf(bitarray.name), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				testPopulateRand(b, bitarray.create, 50000, true)
			}
		})
	}
}

func BenchmarkRandPop100k(b *testing.B) {
	for _, bitarray := range fastBitArrayTypes {
		b.Run(fmt.Sprintf(bitarray.name), func(b *testing.B) {
			var k2 *K2Tree
			for n := 0; n < b.N; n++ {
				k2 = testPopulateRand(b, bitarray.create, 100000, false)
				stats := k2.Stats()
				b.ReportMetric(stats.BitsPerLink, "bits/link")
			}
		})
	}
}

func BenchmarkIncPop50k(b *testing.B) {
	for _, k2config := range smallK2Configs {
		for _, bitarrayt := range testBitArrayTypes {
			b.Run(fmt.Sprint(k2config.name, bitarrayt.name), func(b *testing.B) {
				var k2 *K2Tree
				for n := 0; n < b.N; n++ {
					var err error
					k2, err = newK2Tree(
						bitarrayt.create,
						k2config.config,
					)
					if err != nil {
						b.Fatal(err)
					}
					populateIncrementalTree(50000, k2, false)
				}
				stats := k2.Stats()
				b.ReportMetric(stats.BitsPerLink, "bits/link")
			})
		}

	}
}

func BenchmarkIncPop1M(b *testing.B) {
	for _, bitarrayt := range fastBitArrayTypes {
		b.Run(fmt.Sprint(bitarrayt.name), func(b *testing.B) {
			var k2 *K2Tree
			for n := 0; n < b.N; n++ {
				var err error
				k2, err = newK2Tree(
					bitarrayt.create,
					SixteenFourConfig,
				)
				if err != nil {
					b.Fatal(err)
				}
				populateIncrementalTree(1000000, k2, false)
			}
			stats := k2.Stats()
			b.ReportMetric(stats.BitsPerLink, "bits/link")
			b.ReportMetric(float64(stats.Bytes), "bytes")
		})
	}
}

func BenchmarkIncPop10M(b *testing.B) {
	//b.Skip("IncPop10M runs some stress tests to measure scalability")
	for _, bitarrayt := range tenMillionBitArrayTypes {
		b.Run(fmt.Sprint(bitarrayt.name), func(b *testing.B) {
			var k2 *K2Tree
			for n := 0; n < b.N; n++ {
				var err error
				k2, err = newK2Tree(
					bitarrayt.create,
					SixteenFourConfig,
				)
				if err != nil {
					b.Fatal(err)
				}
				populateIncrementalTree(10000000, k2, true)
			}
			stats := k2.Stats()
			b.ReportMetric(stats.BitsPerLink, "bits/link")
			b.ReportMetric(float64(stats.Bytes), "bytes")
		})
	}
}

func BenchmarkRandPop1M(b *testing.B) {
	//b.Skip("RandPop1M runs some stress tests to measure scalability")
	for _, bitarrayt := range tenMillionBitArrayTypes {
		b.Run(fmt.Sprint(bitarrayt.name), func(b *testing.B) {
			var k2 *K2Tree
			for n := 0; n < b.N; n++ {
				var err error
				k2, err = newK2Tree(
					bitarrayt.create,
					SixteenSixteenConfig,
				)
				if err != nil {
					b.Fatal(err)
				}
				populateRandomTree(1000000, 50000000, k2, true)
			}
			stats := k2.Stats()
			b.ReportMetric(stats.BitsPerLink, "bits/link")
			b.ReportMetric(float64(stats.Bytes), "bytes")
		})
	}
}

func BenchmarkIncPopVar(b *testing.B) {
	for _, k2config := range testK2Configs {
		for _, bitarrayt := range fastBitArrayTypes {
			b.Run(fmt.Sprint(k2config.name, bitarrayt.name), func(b *testing.B) {
				k2, err := newK2Tree(bitarrayt.create, k2config.config)
				if err != nil {
					b.Fatal(err)
				}
				b.ResetTimer()
				populateIncrementalTree(b.N, k2, false)
				stats := k2.Stats()
				b.ReportMetric(stats.BitsPerLink, "bits/link")
			})
		}

	}
}

func BenchmarkRandPopVar(b *testing.B) {
	for _, k2config := range testK2Configs {
		for _, bitarrayt := range fastBitArrayTypes {
			b.Run(fmt.Sprint(k2config.name, bitarrayt.name), func(b *testing.B) {
				k2, err := newK2Tree(bitarrayt.create, k2config.config)
				if err != nil {
					b.Fatal(err)
				}
				b.ResetTimer()
				populateRandomTree(b.N, b.N*2, k2, false)
				stats := k2.Stats()
				b.ReportMetric(stats.BitsPerLink, "bits/link")
			})
		}

	}
}

func BenchmarkRandUnindexed(b *testing.B) {
	for _, bitarrayt := range unindexedBitArrayTypes {
		b.Run(bitarrayt.name, func(b *testing.B) {
			k2, err := newK2Tree(bitarrayt.create, SixteenSixteenConfig)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			populateRandomTree(b.N, b.N*2, k2, false)
			stats := k2.Stats()
			b.ReportMetric(stats.BitsPerLink, "bits/link")
		})
	}
}

func BenchmarkIncUnindexed(b *testing.B) {
	for _, bitarrayt := range unindexedBitArrayTypes {
		b.Run(bitarrayt.name, func(b *testing.B) {
			k2, err := newK2Tree(bitarrayt.create, SixteenSixteenConfig)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			populateIncrementalTree(b.N, k2, false)
			stats := k2.Stats()
			b.ReportMetric(stats.BitsPerLink, "bits/link")
		})
	}
}

var unindexedBitArrayTypes []bitArrayType = []bitArrayType{
	{
		create: func() bitarray {
			return &sliceArray{}
		},
		name: "SliceArray",
	},
	{
		create: func() bitarray {
			return newPagedSliceArray(128 * 1024)
		},
		name: "Paged128kb",
	},
	{
		create: func() bitarray {
			return newByteArray(bytearray.NewSlice())
		},
		name: "ByteArray:Slice",
	},
	{
		create: func() bitarray {
			return newByteArray(bytearray.NewFrontSlice(1024))
		},
		name: "ByteArray:Front",
	},
	{
		create: func() bitarray {
			return newByteArray(bytearray.NewPaged(4096, 0.8, 0.3))
		},
		name: "ByteArray:Paged:4096:80:30",
	},
}

var tenMillionBitArrayTypes []bitArrayType = []bitArrayType{
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newPagedBitarray(1024*128, 0.8, 0.3), 64)
		},
		name: "LRU64Pagedbit128k8030",
	},
}

var fastBitArrayTypes []bitArrayType = []bitArrayType{
	{
		create: func() bitarray {
			return &sliceArray{}
		},
		name: "Root",
	},
	{
		create: func() bitarray {
			return newInt16Index(&sliceArray{})
		},
		name: "Int16",
	},
	{
		create: func() bitarray {
			return newInt16Index(newPagedSliceArray(128 * 1024))
		},
		name: "Int16Paged128kb",
	},
	{
		create: func() bitarray {
			return newInt16Index(newByteArray(bytearray.NewSlice()))
		},
		name: "Int16BASlice",
	},
	{
		create: func() bitarray {
			return newByteArray(bytearray.NewInt16Index(bytearray.NewSlice()))
		},
		name: "BAInt16Slice",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newInt16Index(&sliceArray{}), 64)
		},
		name: "Int16LRU64",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(&sliceArray{}, 64)
		},
		name: "LRU64",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newPagedSliceArray(128*1024), 64)
		},
		name: "LRU64Paged128k",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newPagedBitarray(1024*128, 0.8, 0.3), 64)
		},
		name: "LRU64Pagedbit128k8030",
	},
}

type k2configTest struct {
	config Config
	name   string
}

var smallK2Configs []k2configTest = []k2configTest{
	{
		config: FourFourConfig,
		name:   "4x4",
	},
	{
		config: SixteenFourConfig,
		name:   "16x4",
	},
}

var testK2Configs []k2configTest = []k2configTest{
	{
		config: SixteenFourConfig,
		name:   "16x4",
	},
	{
		config: SixteenSixteenConfig,
		name:   "16x16",
	},
	//{
	//config: SixtySixteenConfig,
	//name:   "64x16",
	//},
	//{
	//config: Config{
	//TreeLayerDef: TwoFiftySixBitsPerLayer,
	//CellLayerDef: SixteenBitsPerLayer,
	//},
	//name: "256x16",
	//},
}
