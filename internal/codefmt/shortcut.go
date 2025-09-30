package codefmt

import (
	"go/ast"
	"go/token"
	"go/types"
	"io"

	"golang.org/x/tools/go/packages"
)

// FormatType is a shorthand for [Formatter.Type].
func FormatType(pkger Pkger, typ types.Type) string {
	return newByPkger(pkger).Type(typ)
}

// FormatTypeParen is a shorthand for [Formatter.TypeParen].
func FormatTypeParen(pkger Pkger, typ types.Type) string {
	return newByPkger(pkger).TypeParen(typ)
}

// FormatObj is a shorthand for [Formatter.Obj].
func FormatObj(pkger Pkger, obj types.Object) string {
	return newByPkger(pkger).Obj(obj)
}

// FormatExpr is a shorthand for [Formatter.Expr].
func FormatExpr(pkger Pkger, expr ast.Expr) string {
	return newByPkger(pkger).Expr(expr)
}

// FormatSig is a shorthand for [Formatter.Sig].
func FormatSig(pkger Pkger, sig *types.Signature) string {
	return newByPkger(pkger).Sig(sig)
}

// FormatPos is a shorthand for [Formatter.Pos].
func FormatPos(pkger Pkger, pos token.Pos) string {
	return newByPkger(pkger).Pos(pos)
}

func Sprintf(pkger Pkger, format string, args ...any) string {
	return newByPkger(pkger).Sprintf(format, args...)
}

func Fprintf(pkger Pkger, w io.Writer, format string, args ...any) (int, error) {
	return newByPkger(pkger).Fprintf(w, format, args...)
}

func Errorf(pkger Pkger, poser Poser, format string, args ...any) error {
	return newByPkger(pkger).Errorf(poser, format, args...)
}

type pkger struct{ pkg *packages.Package }

func (p pkger) Pkg() *packages.Package { return p.pkg }
func Pkg(pkg *packages.Package) Pkger  { return pkger{pkg} }

type poser struct{ pos token.Pos }

func (p poser) Pos() token.Pos { return p.pos }
func Pos(pos token.Pos) Poser  { return poser{pos} }

type typer struct{ typ types.Type }

func (t typer) Type() types.Type { return t.typ }
func Type(typ types.Type) Typer  { return typer{typ} }
