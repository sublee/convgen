package codefmt

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"

	"golang.org/x/tools/go/packages"

	"github.com/sublee/convgen/internal/typeinfo"
)

type (
	Pkger         interface{ Pkg() *packages.Package }
	Poser         interface{ Pos() token.Pos }
	Ender         interface{ End() token.Pos }
	Exprer        interface{ Expr() ast.Expr }
	Objecter      interface{ Object() types.Object }
	Typer         interface{ Type() types.Type }
	TypeInfoer    interface{ TypeInfo() typeinfo.Type }
	TypeInfoTyper interface{ Type() typeinfo.Type }
)

func (f Formatter) wrapPrintfArgs(args []any) []any {
	for i, arg := range args {
		switch arg := arg.(type) {
		case token.Pos, token.Position:
			args[i] = formatArg{arg, f}
		case ast.Expr, types.Object, types.Type:
			args[i] = formatArg{arg, f}
		case Poser, Exprer, Objecter, Typer, TypeInfoer, TypeInfoTyper:
			args[i] = formatArg{arg, f}
		}
	}
	return args
}

type formatArg struct {
	x   any
	fmt Formatter
}

func (f formatArg) Object() types.Object {
	switch x := f.x.(type) {
	case types.Object:
		return x
	case Objecter:
		return x.Object()
	case types.Type:
		if named, ok := x.(*types.Named); ok {
			return named.Obj()
		}
	case Typer:
		if named, ok := x.Type().(*types.Named); ok {
			return named.Obj()
		}
	case TypeInfoer:
		if named, ok := x.TypeInfo().Type().(*types.Named); ok {
			return named.Obj()
		}
	case TypeInfoTyper:
		if named, ok := x.Type().Type().(*types.Named); ok {
			return named.Obj()
		}
	}
	return nil
}

func (f formatArg) Expr() ast.Expr {
	switch x := f.x.(type) {
	case ast.Expr:
		return x
	case Exprer:
		return x.Expr()
	}
	return nil
}

func (f formatArg) Type() types.Type {
	switch x := f.x.(type) {
	case types.Type:
		return x
	case Typer:
		return x.Type()
	case TypeInfoer:
		return x.TypeInfo().Type()
	case TypeInfoTyper:
		return x.Type().Type()
	}
	if obj := f.Object(); obj != nil {
		return obj.Type()
	}
	if expr := f.Expr(); expr != nil {
		return f.fmt.TypesInfo.TypeOf(expr)
	}
	return nil
}

func (f formatArg) Position() *token.Position {
	switch x := f.x.(type) {
	case token.Position:
		return &x
	case token.Pos:
		p := f.fmt.Fset.Position(x)
		return &p
	case Poser:
		p := f.fmt.Fset.Position(x.Pos())
		return &p
	}
	if obj := f.Object(); obj != nil {
		p := f.fmt.Fset.Position(obj.Pos())
		return &p
	}
	return nil
}

// Format implements fmt.Formatter interface.
//
// Supported verbs:
//
//	%o: types.Object (e.g., *types.TypeName, *types.Func) - short form
//	%t: types.Type - short form
//	%q: types.Type - with parentheses for composite types
//	%c: ast.Expr - code form
//	%b: token.Position - file:line:column form
//
// For other verbs, it falls back to the default formatting of fmt package.
func (f formatArg) Format(s fmt.State, verb rune) {
	obj := f.Object()
	expr := f.Expr()
	typ := f.Type()
	pos := f.Position()
	switch verb {
	case 'o':
		if obj == nil {
			fmt.Fprintf(s, "[%%o cannot format %T]", f.x)
			return
		}
		_, _ = s.Write([]byte(f.fmt.Obj(obj)))

	case 't':
		if typ == nil {
			fmt.Fprintf(s, "[%%t cannot format %T]", f.x)
			return
		}
		_, _ = s.Write([]byte(f.fmt.Type(typ)))

	case 'q':
		if typ == nil {
			fmt.Fprintf(s, "[%%q cannot format %T]", f.x)
			return
		}
		_, _ = s.Write([]byte(f.fmt.TypeParen(typ)))

	case 'c':
		if expr == nil {
			fmt.Fprintf(s, "[%%c cannot format %T]", f.x)
			return
		}
		_, _ = s.Write([]byte(f.fmt.Expr(expr)))

	case 'b':
		if pos == nil {
			fmt.Fprintf(s, "[%%b cannot format %T]", f.x)
			return
		}
		_, _ = s.Write([]byte(FormatPosition(*pos)))

	default:
		fmt.Fprintf(s, fmt.FormatString(s, verb), f.x)
	}
}

func (f Formatter) Sprintf(format string, args ...any) string {
	args = f.wrapPrintfArgs(args)
	return fmt.Sprintf(format, args...)
}

func (f Formatter) Fprintf(w io.Writer, format string, args ...any) (int, error) {
	args = f.wrapPrintfArgs(args)
	return fmt.Fprintf(w, format, args...)
}
