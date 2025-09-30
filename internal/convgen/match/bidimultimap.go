package match

import (
	"iter"

	"github.com/emirpasic/gods/maps/linkedhashmap"
	"github.com/emirpasic/gods/sets/linkedhashset"
)

type bidiMultiMap[K, V comparable] struct {
	fwd *linkedhashmap.Map // key: K, value: *linkedhashset.Set of V
	bwd *linkedhashmap.Map // key: V, value: *linkedhashset.Set of K
}

func newBidiMultiMap[K, V comparable]() *bidiMultiMap[K, V] {
	return &bidiMultiMap[K, V]{
		fwd: linkedhashmap.New(),
		bwd: linkedhashmap.New(),
	}
}

func (m *bidiMultiMap[K, V]) Has(k K, v V) bool {
	vs, ok := m.fwd.Get(k)
	if !ok {
		return false
	}
	return vs.(*linkedhashset.Set).Contains(v)
}

func (m *bidiMultiMap[K, V]) Add(k K, v V) {
	vs, ok := m.fwd.Get(k)
	if !ok {
		vs = linkedhashset.New()
		m.fwd.Put(k, vs)
	}
	vs.(*linkedhashset.Set).Add(v)

	ks, ok := m.bwd.Get(v)
	if !ok {
		ks = linkedhashset.New()
		m.bwd.Put(v, ks)
	}
	ks.(*linkedhashset.Set).Add(k)
}

func (m *bidiMultiMap[K, V]) Delete(k K, v V) {
	vs, ok := m.fwd.Get(k)
	if ok {
		vs.(*linkedhashset.Set).Remove(v)
		if vs.(*linkedhashset.Set).Size() == 0 {
			m.fwd.Remove(k)
		}
	}

	ks, ok := m.bwd.Get(v)
	if ok {
		ks.(*linkedhashset.Set).Remove(k)
		if ks.(*linkedhashset.Set).Size() == 0 {
			m.bwd.Remove(v)
		}
	}
}

func (m *bidiMultiMap[K, V]) Get(k K) []V {
	vset, ok := m.fwd.Get(k)
	if !ok {
		return nil
	}

	var vs []V
	for it := vset.(*linkedhashset.Set).Iterator(); it.Next(); {
		vs = append(vs, it.Value().(V))
	}
	return vs
}

func (m *bidiMultiMap[K, V]) GetKeys(v V) []K {
	kset, ok := m.bwd.Get(v)
	if !ok {
		return nil
	}

	var ks []K
	for it := kset.(*linkedhashset.Set).Iterator(); it.Next(); {
		ks = append(ks, it.Value().(K))
	}
	return ks
}

func (m *bidiMultiMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for it := m.fwd.Iterator(); it.Next(); {
			k := it.Key().(K)
			vs := it.Value().(*linkedhashset.Set)
			for it := vs.Iterator(); it.Next(); {
				if !yield(k, it.Value().(V)) {
					return
				}
			}
		}
	}
}
