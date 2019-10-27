package k2tree

import (
	"encoding/binary"
	"errors"
	"os"

	mmap "github.com/barakmich/mmap-go"
)

type pagefile struct {
	bytes    mmap.MMap
	file     *os.File
	filelen  int
	pages    int
	pagesize int
}

/*
File Layout:

----0----------------------
Header
	MagicNo (8B)
	PageSize (8B/int64)
	N-Pages (8B/int64)
	PagefileDataReserve
	----32KiB------------------
	PagefileUserMetadata
----128KiB-----------------
Page0
----128KiB + PageSize------
Page1
----128KiB + 2*PageSize----
Page2
....

Page Layout:

----0----------------------
PageHeader
----4KiB-------------------
PageBytes
----PageSize---------------

*/

const (
	// blockSize is the page size for most OSes, and block size for most FSes
	// We'll refer to them as blocks in this code. We want all things to be
	// multiples of this size (in bytes).
	blockSize          = 4096
	headerSize         = 128 * 1024
	userMetadataOffset = 32 * 1024
	DefaultPagesize    = 512 * 1024
)

var (
	// magicHeader identifies the file and are the first bytes written in the header
	// page. The whole of the header is 128KiB, most of which (96KiB) is for users of the page file
	// (ie, the K2 bitarray/pagefile interface) and the front part of which is for bookkeeping in the pagefile itself.
	// It also identifies the version. The versions are immutable, and a converter
	// must be run if it changes.
	magicHeader = []byte{'K', '2', 'B', 'P', 'v', '1', 0x00, 0x00}
)

type header struct {
	Magic    [8]byte
	PageSize int64
	Pages    int64
}

func createPagefile(filename string, pagesize int) (*pagefile, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	err = f.Truncate(headerSize)
	if err != nil {
		return nil, err
	}
	var h header
	h.PageSize = int64(pagesize)
	writeHeader(h, f)
	err = f.Close()
	if err != nil {
		return nil, err
	}
	return openPagefile(filename)
}

func writeHeader(h header, f *os.File) error {
	for i, b := range magicHeader {
		h.Magic[i] = b
	}
	f.Seek(0, 0)
	return binary.Write(f, binary.BigEndian, &h)
}

func (s *pagefile) Close() error {
	err := s.bytes.Unmap()
	if err != nil {
		return err
	}
	var h header
	h.PageSize = int64(s.pagesize)
	h.Pages = int64(s.pages)
	err = writeHeader(h, s.file)
	if err != nil {
		return err
	}
	return s.file.Close()
}

func openPagefile(filename string) (*pagefile, error) {
	f, err := os.OpenFile(filename, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	var h header
	err = binary.Read(f, binary.BigEndian, &h)
	if err != nil {
		return nil, err
	}
	for i, b := range magicHeader {
		if b != h.Magic[i] {
			return nil, errors.New("incompatible magic header")
		}
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	f.Seek(0, 0)
	m, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		return nil, err
	}
	return &pagefile{
		bytes:    m,
		pages:    int(h.Pages),
		pagesize: int(h.PageSize),
		file:     f,
		filelen:  int(fi.Size()),
	}, nil
}

func newPagefile(filename string, pagesize int) (*pagefile, error) {
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return createPagefile(filename, pagesize)
		}
		return nil, err
	}
	return openPagefile(filename)
}
