package k2tree

import (
	"fmt"
	"math/rand"
	"time"
)

func populateRandomTree(nLinks, maxID int, k2 *K2Tree) (maxrow int, maxcol int) {
	//fmt.Println("Populating Tree...")
	rowcnt := make(map[int]int)
	colcnt := make(map[int]int)

	for i := 0; i < nLinks; i++ {
		if i%10000 == 0 {
			//		fmt.Println(i)
		}
		row := rand.Intn(maxID)
		col := rand.Intn(maxID)
		k2.Add(row, col)
		rowcnt[row]++
		colcnt[col]++
	}

	maxrowcnt := 0
	for k, v := range rowcnt {
		if v > maxrowcnt {
			maxrow = k
		}
	}

	maxcolcnt := 0
	for k, v := range colcnt {
		if v > maxcolcnt {
			maxcol = k
		}
	}
	return
}

func populateIncrementalTree(nLinks int, k2 *K2Tree, debug bool) (maxrow int, maxcol int) {
	//fmt.Println("Populating Tree...")
	rowcnt := make(map[int]int)
	colcnt := make(map[int]int)
	var row int
	var col int

	timebefore := time.Now()

	for i := 0; i < nLinks; i++ {
		if debug && i%100000 == 0 {
			timeafter := time.Now()
			td := timeafter.Sub(timebefore)
			timebefore = timeafter
			fmt.Println(i, td)
		}
		rowd := rand.Intn(10)
		cold := rand.Intn(10)
		rowd = rowd - 3
		row = row + rowd
		if row < 0 {
			row = 0
		}
		cold = cold - 5
		col = col + cold
		if col < 0 {
			col = 0
		}
		k2.Add(row, col)
	}

	maxrowcnt := 0
	for k, v := range rowcnt {
		if v > maxrowcnt {
			maxrow = k
		}
	}

	maxcolcnt := 0
	for k, v := range colcnt {
		if v > maxcolcnt {
			maxcol = k
		}
	}
	return
}
