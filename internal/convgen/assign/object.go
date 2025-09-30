package assign

import (
	"go/token"

	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/typeinfo"
)

// Object is the common interface for things that can be assigned to another
// things. It must provide the type information and may have the name. Here is
// the list of objects:
//
// - pointer
// - map key/value
// - slice element
// - struct field
// - union member
// - enum member
type Object interface {
	// With the following struct definitions in example.com/pkg package:
	//	type Session struct {
	//		SignedUser User
	//	}
	//	type User struct {
	//		Name string
	//	}
	// The following shows the values of the methods for the field
	// Session.SignedUser.Name:
	Type() typeinfo.Type // e.g., string
	QualName() string    // e.g., User.Name
	CrumbName() string   // e.g., Session.SignedUser.Name
	DebugName() string   // e.g., Session.SignedUser.Name (type string)
	Exported() bool
	Pos() token.Pos
}

// anonObject is an anonymous object used for type-only objects or element
// objects.
type anonObject struct {
	t   typeinfo.Type
	qn  string
	cn  string
	dn  string
	e   bool
	pos token.Pos
}

func (o anonObject) Type() typeinfo.Type { return o.t }
func (o anonObject) QualName() string    { return o.qn }
func (o anonObject) CrumbName() string   { return o.cn }
func (o anonObject) DebugName() string   { return o.dn }
func (o anonObject) Exported() bool      { return o.e }
func (o anonObject) Pos() token.Pos      { return o.pos }

// typeOnly returns an object that has only type information without the name.
func typeOnly(pkger codefmt.Pkger, t typeinfo.Type) Object {
	typeName := codefmt.FormatType(pkger, t.T)
	exported := false
	if t.IsNamed() {
		exported = t.Named.Obj().Exported()
	}
	return anonObject{t, typeName, typeName, typeName, exported, t.Pos()}
}

// elemOf returns an object of the element type of the given object. Pointers,
// maps, slices, and arrays all have element types.
func elemOf(o Object) Object {
	return anonObject{*o.Type().Elem, o.QualName(), o.CrumbName(), o.DebugName(), o.Exported(), o.Pos()}
}
