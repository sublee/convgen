package typeinfo

import (
	"fmt"
	"go/token"
	"go/types"
)

// Type describes a type information. It holds information of [types.Type] that
// is necessary from the Convgen's perspective.
type Type struct {
	T types.Type

	Basic     *types.Basic
	Array     *types.Array
	Slice     *types.Slice
	Map       *types.Map
	Struct    *types.Struct
	Interface *types.Interface
	Pointer   *types.Pointer
	Named     *types.Named

	Elem *Type
	Key  *Type
	Len  int64
}

func (t Type) Type() types.Type { return t.T }
func (t Type) String() string   { return t.T.String() }

func (t Type) IsBasic() bool     { return t.Basic != nil }
func (t Type) IsArray() bool     { return t.Array != nil }
func (t Type) IsSlice() bool     { return t.Slice != nil }
func (t Type) IsMap() bool       { return t.Map != nil }
func (t Type) IsStruct() bool    { return t.Struct != nil }
func (t Type) IsInterface() bool { return t.Interface != nil }
func (t Type) IsPointer() bool   { return t.Pointer != nil }
func (t Type) IsNamed() bool     { return t.Named != nil }

func (t Type) IsNil() bool   { return t.T == types.Universe.Lookup("nil").Type() }
func (t Type) IsError() bool { return t.T == types.Universe.Lookup("error").Type() }

func (t Type) Identical(u Type) bool { return types.Identical(t.T, u.T) }

// TypeOf inspects the given type and returns a new [Type].
func TypeOf(t types.Type) Type {
	switch tt := types.Unalias(t).(type) {
	case *types.Basic:
		return Type{T: t, Basic: tt}
	case *types.Array:
		elem := TypeOf(tt.Elem())
		return Type{T: t, Array: tt, Elem: &elem, Len: tt.Len()}
	case *types.Slice:
		elem := TypeOf(tt.Elem())
		return Type{T: t, Slice: tt, Elem: &elem}
	case *types.Map:
		elem := TypeOf(tt.Elem())
		key := TypeOf(tt.Key())
		return Type{T: t, Map: tt, Elem: &elem, Key: &key}
	case *types.Struct:
		return Type{T: t, Struct: tt}
	case *types.Interface:
		return Type{T: t, Interface: tt}
	case *types.Pointer:
		elem := TypeOf(tt.Elem())
		return Type{T: t, Pointer: tt, Elem: &elem}
	case *types.Named:
		info := TypeOf(tt.Underlying())
		info.T = t
		info.Named = tt
		return info
	case *types.Signature:
		return Type{T: t}
	case *types.Tuple:
		if tt.Len() == 0 {
			return Type{T: t}
		}
	}
	panic(fmt.Errorf("unknown type: %T", t))
}

// Pkg returns the package where the type is defined. It returns nil if the type
// is not a named type.
func (t Type) Pkg() *types.Package {
	if !t.IsNamed() {
		return nil
	}
	return t.Named.Obj().Pkg()
}

// Pos returns the position where the type is defined. It returns token.NoPos if
// the type is not a named type.
func (t Type) Pos() token.Pos {
	if t.IsNamed() {
		return t.Named.Obj().Pos()
	}
	if t.IsPointer() {
		return t.Deref().Pos()
	}
	return token.NoPos
}

// Ref returns the pointer type of the type. For type of X, it returns type of
// *X.
func (t Type) Ref() Type {
	return TypeOf(types.NewPointer(t.T))
}

// Deref returns the element type if the type is a pointer. For type of *X, it
// returns type of X. If the type is not a pointer, it returns the type itself.
func (t Type) Deref() Type {
	if t.IsPointer() {
		return (*t.Elem).Deref()
	}
	return t
}

// PointerDepth returns the number of pointer indirections. For example, for
// type of ***X, it returns 3. For type of X, it returns 0.
func (t Type) PointerDepth() int {
	depth := 0
	for t.IsPointer() {
		depth++
		t = *t.Elem
	}
	return depth
}

// Method returns the method with the given name. If the type is not a named
// type or the method does not exist, it returns nil and false.
func (t Type) Method(name string) (*types.Func, bool) {
	if !t.IsNamed() {
		return nil, false
	}

	for method := range t.Named.Methods() {
		if method.Name() == name {
			return method, true
		}
	}

	return nil, false
}

// StructField returns the struct field with the given name. If the type is not
// a struct or the field does not exist, it returns nil and false.
func (t Type) StructField(name string) (*types.Var, bool) {
	if !t.IsStruct() {
		return nil, false
	}

	for field := range t.Struct.Fields() {
		if field.Name() == name {
			return field, true
		}
	}

	return nil, false
}

// IsGeneric reports whether the type is generic or has any generic type
// parameters. Even though the type has type parameters, if all type arguments
// are concrete types, it returns false.
func (t Type) IsGeneric() bool {
	return isGeneric(t.T)
}

func isGeneric(t types.Type) bool {
	switch t := types.Unalias(t).(type) {
	case *types.Named:
		if t.TypeParams().Len() == 0 {
			// No type parameters
			// e.g., Foo
			return false
		}

		targs := t.TypeArgs()
		if targs.Len() == 0 {
			// Have type parameters but no arguments
			// e.g., Foo[T]
			return true
		}

		for i := 0; i < targs.Len(); i++ {
			if isGeneric(targs.At(i)) {
				// Some type argument is generic
				// e.g., Foo[int, T]
				return true
			}
		}
	case *types.Struct:
		for f := range t.Fields() {
			if isGeneric(f.Type()) {
				return true
			}
		}
	case *types.Interface:
		for m := range t.Methods() {
			if isGeneric(m.Type()) {
				return true
			}
		}
	case *types.Signature:
		return t.TypeParams().Len() != 0
	case *types.TypeParam:
		return true
	}
	return false
}
