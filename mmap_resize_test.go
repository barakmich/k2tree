package k2tree

import (
	"os"
	"testing"

	mmap "github.com/barakmich/mmap-go"
)

func TestMmapGrow(t *testing.T) {
	f, err := os.Create("mmap_resize_testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("mmap_resize_testfile")
	defer f.Close()
	err = f.Truncate(3)
	if err != nil {
		t.Fatal(err)
	}
	m, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Unmap()
	m[0] = 'a'
	m[1] = 'b'
	m[2] = 'c'
	err = m.Truncate(f, 4)

	if err != nil {
		t.Fatal(err)
	}
	m[3] = 'd'
	err = m.Truncate(f, 2)
	if err != nil {
		t.Fatal(err)
	}
}
