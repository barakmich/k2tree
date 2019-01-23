package k2tree

type bitarray interface {
	Len() int
	Set(at int, val bool)
	Get(at int) bool
	Count(from, to int) int
	Total() int
	Insert(n int, at int) error
	debug() string
}
