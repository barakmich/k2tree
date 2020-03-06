package bytearray

import (
	"fmt"
	"testing"
)

func BenchmarkInsertPatternArray(b *testing.B) {
	for _, arraytype := range arrayTypes {
		b.Run(fmt.Sprintf(arraytype.name), func(b *testing.B) {
			b.ReportAllocs()
			tv := insertTestVector()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				vec := arraytype.makeArray()
				for i := 0; i < 4; i++ {
					for _, x := range tv {
						vec.Insert(x, []byte{0x01, 0x01})
					}
				}
			}

		})
	}
}

func BenchmarkSpilloverMatrix(b *testing.B) {
	highwaters := []float64{0.75, 0.8, 0.9}
	lowwaters := []float64{0.3, 0.5, 0.7}
	multipliers := []bool{false, true}
	pagesizes := []int{1024, 4 * 1024}
	for _, p := range pagesizes {
		for _, m := range multipliers {
			for _, h := range highwaters {
				for _, l := range lowwaters {
					b.Run(fmt.Sprintf("%d:%.2f:%.2f:%v", p, h, l, m), func(b *testing.B) {
						b.ReportAllocs()
						tv := insertTestVector()
						b.ResetTimer()
						for n := 0; n < b.N; n++ {
							vec := NewSpillover(p, h, l, m)
							for i := 0; i < 10; i++ {
								for _, x := range tv {
									vec.Insert(x, []byte{0x01, 0x01})
								}
							}
						}

					})
				}
			}
		}
	}
}

type testarraytype struct {
	makeArray func() ByteArray
	name      string
}

var arrayTypes []testarraytype = []testarraytype{
	{
		makeArray: func() ByteArray {
			return &SliceArray{}
		},
		name: "Slice:::::::::",
	},
	{
		makeArray: func() ByteArray {
			return NewFrontSlice(1024)
		},
		name: "FrontSlice1k:::",
	},
	{
		makeArray: func() ByteArray {
			return NewFrontSlice(1024 * 1024)
		},
		name: "FrontSlice1M:::",
	},
	{
		makeArray: func() ByteArray {
			return NewSpillover(1024, 0.8, 0.5, true)
		},
		name: "Spillover::1024:80:50:2x",
	},
	{
		makeArray: func() ByteArray {
			return NewSpillover(4096, 0.8, 0.3, false)
		},
		name: "Spillover::4096:80:30:1x",
	},
	{
		makeArray: func() ByteArray {
			return NewSpillover(32*1024, 0.8, 0.3, false)
		},
		name: "Spillover::32k:80:30:1x",
	},
}
