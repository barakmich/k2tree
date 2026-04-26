package k2tree

import (
	"math/rand"
	"sort"
	"testing"
)

func FuzzIteratorEdges(f *testing.F) {
	f.Fuzz(func(t *testing.T, seed uint64) {
		// Seed a reproducible random source from the fuzz input.
		rng := rand.New(rand.NewSource(int64(seed)))

		maxID := 2048          // node indices in range [0, 2047]
		nEdges := 200 + rng.Intn(300) // 200–499 edges per fuzz run

		// Build a K2Tree using the simple slice-backed bitarray.
		k2, err := newK2Tree(func() bitarray { return &sliceArray{} }, DefaultConfig)
		if err != nil {
			t.Fatal(err)
		}

		// Track ground truth: which edges (row, col) were inserted.
		// rowToCols[row] is a map of all unique columns for that row.
		rowToCols := make(map[int]map[int]bool)
		colToRows := make(map[int]map[int]bool)

		for i := 0; i < nEdges; i++ {
			row := rng.Intn(maxID)
			col := rng.Intn(maxID)

			k2.Add(row, col)

			if rowToCols[row] == nil {
				rowToCols[row] = make(map[int]bool)
			}
			rowToCols[row][col] = true

			if colToRows[col] == nil {
				colToRows[col] = make(map[int]bool)
			}
			colToRows[col][row] = true
		}

		// Verify row iterators: for each row with outgoing edges,
		// the iterator must return exactly the expected columns.
		rows := collectKeys(rowToCols)
		for _, row := range rows {
			expectedCols := mapToSortedSlice(rowToCols[row])

			it := k2.From(row)
			actualCols := it.ExtractAll()
			sort.Ints(actualCols)

			if !intSlicesEqual(actualCols, expectedCols) {
				t.Fatalf("row iterator mismatch for row %d (seed=%d):\n  actual   = %v\n  expected = %v",
					row, seed, actualCols, expectedCols)
			}
		}

		// Verify column iterators: for each column with incoming edges,
		// the iterator must return exactly the expected rows.
		cols := collectKeys(colToRows)
		for _, col := range cols {
			expectedRows := mapToSortedSlice(colToRows[col])

			it := k2.To(col)
			actualRows := it.ExtractAll()
			sort.Ints(actualRows)

			if !intSlicesEqual(actualRows, expectedRows) {
				t.Fatalf("col iterator mismatch for col %d (seed=%d):\n  actual   = %v\n  expected = %v",
					col, seed, actualRows, expectedRows)
			}
		}
	})
}

func intSlicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func collectKeys(m map[int]map[int]bool) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func mapToSortedSlice(m map[int]bool) []int {
	slice := make([]int, 0, len(m))
	for k := range m {
		slice = append(slice, k)
	}
	sort.Ints(slice)
	return slice
}