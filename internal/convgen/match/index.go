package match

import (
	"fmt"
	"go/token"
	"strings"

	"github.com/emirpasic/gods/maps/linkedhashmap"
)

// indexer holds entries to be indexed.
type indexer[T object] struct {
	m *linkedhashmap.Map // token.Pos -> entry
}

// newIndexer creates a new indexer.
func newIndexer[T object]() *indexer[T] {
	return &indexer[T]{m: linkedhashmap.New()}
}

// add registers a new value with the given path and key as an entry.
func (i *indexer[T]) add(obj T, key string) {
	if key == "" {
		panic("key must not be empty")
	}
	if strings.Contains(key, ".") {
		panic("key must not contain dots")
	}
	i.m.Put(obj.Pos(), entry{obj, key})
}

func (i *indexer[T]) delete(pos token.Pos) {
	i.m.Remove(pos)
}

// String returns a string representation of the indexer.
func (i *indexer[T]) String() string {
	return strings.SplitN(i.m.String(), "\n", 2)[1]
}

// index holds an index of entries by their paths and keys.
type index struct {
	All   []entry
	ByKey map[string][]entry
	ByPos map[token.Pos]entry
}

// build builds an index of entries by their paths and keys.
func (i *indexer[T]) build(renamers []renameFunc, commonFinders []findCommonFunc) index {
	idx := index{
		All:   make([]entry, 0, i.m.Size()),
		ByKey: make(map[string][]entry, i.m.Size()/2),
		ByPos: make(map[token.Pos]entry, i.m.Size()),
	}

	it := i.m.Iterator()
	for it.Next() {
		e := it.Value().(entry)
		idx.All = append(idx.All, e)
	}

	// rename keys
	for i, rename := range renamers {
		keys := make([]string, len(idx.All))
		for i, e := range idx.All {
			keys[i] = e.key
		}

		var common string
		if find := commonFinders[i]; find != nil && len(keys) > 1 {
			common = find(keys)
		}

		for j := range idx.All {
			idx.All[j].key = rename(idx.All[j].key, common)
		}
	}

	for _, e := range idx.All {
		idx.ByPos[e.Pos()] = e
		idx.ByKey[e.key] = append(idx.ByKey[e.key], e)
	}

	return idx
}

// links maintains bidirectional links between two sets of entries.
type links struct {
	m *bidiMultiMap[entry, entry]
}

// newLinks creates a new links.
func newLinks() *links {
	return &links{newBidiMultiMap[entry, entry]()}
}

func (l *links) String() string {
	var ss []string
	for x, y := range l.m.All() {
		ss = append(ss, fmt.Sprintf("%s -> %s", x, y))
	}
	return strings.Join(ss, "\n")
}

// Link links x and y.
func (l *links) Link(x, y entry) {
	l.m.Add(x, y)
}

// Unlink unlinks x and y.
func (l *links) Unlink(x, y entry) {
	l.m.Delete(x, y)
}

// Linked reports whether x and y are linked.
func (l *links) Linked(x, y entry) bool {
	if !x.IsValid() && !y.IsValid() {
		// missing -> missing
		return true
	}
	if !x.IsValid() {
		// missing -> y
		return l.m.Has(missing, y) || len(l.m.GetKeys(y)) == 0
	}
	if !y.IsValid() {
		// x -> missing
		return l.m.Has(x, missing) || len(l.m.Get(x)) == 0
	}
	// x -> y
	return l.m.Has(x, y)
}

// FromX returns all y entries linked from x.
func (l *links) FromX(x entry) []entry {
	return l.m.Get(x)
}

// FromY returns all x entries linked from y.
func (l *links) FromY(y entry) []entry {
	return l.m.GetKeys(y)
}

// entry is a node of an index.
type entry struct {
	object

	// key is the key of the entry. It is usually a short name without dots. The
	// matcher will match entries by their keys rather than their paths.
	// However, the keys may be renamed by renamers before matching.
	key string
}

// IsValid reports whether the entry is valid. An entry is valid if it has an
// object and non-empty key.
func (e entry) IsValid() bool {
	return e.object != nil && e.key != ""
}

func (e entry) Pos() token.Pos {
	if e.object == nil {
		return token.NoPos
	}
	return e.object.Pos()
}

func (e entry) String() string {
	if !e.IsValid() {
		return "?"
	}

	name := e.CrumbName()
	if e.key == name {
		return name
	}
	return fmt.Sprintf("%s [%s]", e.key, name)
}

// missing is a sentinel entry representing a missing entry.
var missing = entry{}
