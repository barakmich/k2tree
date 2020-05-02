package k2tree

import "testing"

func TestPrintTree(t *testing.T) {
	k2, err := New()
	if err != nil {
		t.Fatal(err)
	}
	populateIncrementalTree(10, k2, false)
	k2.printTree()
}
