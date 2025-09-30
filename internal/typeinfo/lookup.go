package typeinfo

import (
	"iter"

	"golang.org/x/tools/go/types/typeutil"
)

// Lookup indexes converter functions by their input and output types.
type Lookup[T Func] struct {
	mapY   *typeutil.Map
	hasher typeutil.Hasher
}

// NewLookup creates a new [Lookup].
func NewLookup[T Func]() *Lookup[T] {
	hasher := typeutil.MakeHasher()
	mapY := new(typeutil.Map)
	mapY.SetHasher(hasher)
	return &Lookup[T]{mapY, hasher}
}

// Put adds a converter function to the registry.
func (l *Lookup[T]) Put(fn T) (T, bool) {
	x, y := fn.X(), fn.Y()

	mapX, ok := l.mapY.At(y.Type()).(*typeutil.Map)
	if !ok {
		mapX = new(typeutil.Map)
		mapX.SetHasher(l.hasher)
		l.mapY.Set(y.Type(), mapX)
	}

	if old, ok := mapX.At(x.Type()).(T); ok {
		return old, false
	}

	if old := mapX.Set(x.Type(), fn); old != nil {
		panic("unexpected old value")
	}
	return *new(T), true
}

// Get finds a function which converts X to Y.
func (l *Lookup[T]) Get(x, y Type) (T, bool) {
	if l == nil {
		return *new(T), false
	}

	mapX, ok := l.mapY.At(y.Type()).(*typeutil.Map)
	if !ok {
		return *new(T), false
	}

	fn, ok := mapX.At(x.Type()).(T)
	if !ok {
		return *new(T), false
	}

	return fn, true
}

// Del removes a function which converts X to Y. It returns whether such a
// function existed.
func (l *Lookup[T]) Del(x, y Type) bool {
	if l == nil {
		return false
	}

	mapX, ok := l.mapY.At(y.Type()).(*typeutil.Map)
	if !ok {
		return false
	}

	return mapX.Delete(x.Type())
}

// Range iterates all registered functions.
func (l *Lookup[T]) Range() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, y := range l.mapY.Keys() {
			mapX := l.mapY.At(y).(*typeutil.Map)
			for _, x := range mapX.Keys() {
				fn, ok := mapX.At(x).(T)
				if !ok {
					continue
				}
				if !yield(fn) {
					return
				}
			}
		}
	}
}
