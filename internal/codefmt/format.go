package codefmt

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Formatter formats types, objects, expressions, and positions.
type Formatter struct {
	PkgPath   string
	Fset      *token.FileSet
	TypesInfo *types.Info
}

func New(pkg *packages.Package) Formatter {
	if pkg == nil {
		return Formatter{}
	}
	return Formatter{pkg.PkgPath, pkg.Fset, pkg.TypesInfo}
}

func newByPkger(pkger Pkger) Formatter {
	if pkger == nil {
		return New(nil)
	}
	return New(pkger.Pkg())
}

// qf is a [types.Qualifier] for types.ObjectString and types.TypeString.
func (f Formatter) qf(pkg *types.Package) string {
	if pkg.Path() == f.PkgPath {
		return ""
	}
	return pkg.Name()
}

// Type returns a string representation of the given type.
//
// e.g., f.Type([types.Type for bytes.Buffer]) => "bytes.Buffer"
func (f Formatter) Type(typ types.Type) string {
	return types.TypeString(typ, f.qf)
}

// TypeParen returns a string representation of the given type. It wraps the
// string with parentheses if the type is a pointer.
func (f Formatter) TypeParen(typ types.Type) string {
	s := f.Type(typ)
	if strings.HasPrefix(s, "*") {
		return fmt.Sprintf("(%s)", s)
	}
	return s
}

// Obj returns a code string to refer the given object.
//
// e.g., f.Obj([types.Object for strconv.Atoi]) => "strconv.Atoi"
func (f Formatter) Obj(obj types.Object) string {
	var b strings.Builder

	if fn, ok := obj.(*types.Func); ok {
		if recv := fn.Signature().Recv(); recv != nil {
			b.WriteString(f.TypeParen(recv.Type()))
			b.WriteByte('.')
		}
	}

	if b.Len() == 0 {
		if pkg := f.qf(obj.Pkg()); pkg != "" {
			b.WriteString(pkg)
			b.WriteByte('.')
		}
	}

	b.WriteString(obj.Name())
	return b.String()
}

// Expr returns a Go source code representation of the given [ast.Expr].
func (f Formatter) Expr(expr ast.Expr) string {
	var b strings.Builder
	if err := format.Node(&b, f.Fset, expr); err != nil {
		panic(err) // should never happen because ast.Expr must be supported by the go/printer
	}
	return b.String()
}

// Sig returns a compact string representation of the given function signature
// without "func" keyword, receiver, and parameter names.
//
// e.g., f.Sig([*types.Signature of strconv.Atoi) => "(string) (int, error)"
func (f Formatter) Sig(sig *types.Signature) string {
	if sig == nil {
		return "<nil>"
	}

	var b strings.Builder

	b.WriteString("(")
	for i := 0; i < sig.Params().Len(); i++ {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(f.Type(sig.Params().At(i).Type()))
	}
	b.WriteString(")")

	switch sig.Results().Len() {
	case 0:
		return b.String()

	case 1:
		b.WriteString(" ")
		b.WriteString(f.Type(sig.Results().At(0).Type()))
		return b.String()

	default:
		b.WriteString(" (")
		for i := 0; i < sig.Results().Len(); i++ {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(f.Type(sig.Results().At(i).Type()))
		}
		b.WriteString(")")
		return b.String()
	}
}

func (f Formatter) Pos(pos token.Pos) string {
	return FormatPosition(f.Fset.Position(pos))
}

// wd is the cached working directory.
var wd, _ = os.Getwd()

func FormatPosition(pos token.Position) string {
	if !pos.IsValid() {
		return "-:-"
	}

	filename := pos.Filename
	if rel, err := filepath.Rel(wd, filename); err == nil {
		filename = rel
	}

	return fmt.Sprintf("%s:%d:%d", filename, pos.Line, pos.Column)
}
