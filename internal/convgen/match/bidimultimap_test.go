package match

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBidiMultiMap_Empty(t *testing.T) {
	m := newBidiMultiMap[int, string]()
	assert.Empty(t, m.Get(42))
	assert.Empty(t, m.GetKeys("test"))
}

func TestBidiMultiMap_Forward(t *testing.T) {
	m := newBidiMultiMap[int, string]()
	m.Add(1, "a")
	m.Add(1, "b")
	m.Add(1, "c")
	assert.Equal(t, []string{"a", "b", "c"}, m.Get(1))
}

func TestBidiMultiMap_Backward(t *testing.T) {
	m := newBidiMultiMap[int, string]()
	m.Add(1, "a")
	m.Add(2, "a")
	m.Add(3, "a")
	assert.Equal(t, []int{1, 2, 3}, m.GetKeys("a"))
}

func TestBidiMultiMap_Has(t *testing.T) {
	m := newBidiMultiMap[int, string]()
	m.Add(1, "a")
	m.Add(1, "b")
	m.Add(2, "c")
	m.Add(2, "d")
	assert.True(t, m.Has(1, "a"))
	assert.True(t, m.Has(1, "b"))
	assert.True(t, m.Has(2, "c"))
	assert.True(t, m.Has(2, "d"))
	assert.False(t, m.Has(1, "c"))
	assert.False(t, m.Has(2, "a"))
	assert.False(t, m.Has(3, "z"))
}

func TestBidiMultiMap_Delete(t *testing.T) {
	m := newBidiMultiMap[int, string]()
	m.Add(1, "a")
	m.Add(1, "b")
	m.Add(2, "c")
	m.Add(2, "d")
	m.Delete(1, "a")
	assert.Len(t, m.Get(1), 1)
	assert.Len(t, m.Get(2), 2)
	assert.Len(t, m.GetKeys("a"), 0)
	assert.Len(t, m.GetKeys("b"), 1)
	assert.Len(t, m.GetKeys("c"), 1)
	assert.Len(t, m.GetKeys("d"), 1)
}

func TestBidiMultiMap_All(t *testing.T) {
	m := newBidiMultiMap[int, string]()
	m.Add(1, "a")
	m.Add(1, "b")
	m.Add(2, "c")
	m.Add(2, "d")
	m.Add(999, "A")
	m.Add(888, "C")

	var ks []int
	var vs []string
	for k, v := range m.All() {
		ks = append(ks, k)
		vs = append(vs, v)
	}

	assert.Equal(t, []int{1, 1, 2, 2, 999, 888}, ks)
	assert.Equal(t, []string{"a", "b", "c", "d", "A", "C"}, vs)
}
