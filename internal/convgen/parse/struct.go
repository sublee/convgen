package parse

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/types/typeutil"

	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/typeinfo"
)

type structParsers struct{ x, y typeinfo.Type }

func (ps structParsers) ParsePathX(p *Parser, expr ast.Expr) (*Path, error) {
	return ps.parse(p, expr, ps.x)
}

func (ps structParsers) ParsePathY(p *Parser, expr ast.Expr) (*Path, error) {
	return ps.parse(p, expr, ps.y)
}

func (ps structParsers) parse(p *Parser, expr ast.Expr, owner typeinfo.Type) (*Path, error) {
	path, err := ps.walk(p, expr, owner)
	if err != nil {
		return nil, err
	}

	if len(path.StructField) == 0 {
		panic("unreachable: struct field must be at least one")
	}
	if len(path.StructField) == 1 {
		return nil, codefmt.Errorf(p, expr, "cannot use %t itself as a field", owner)
	}
	return path, nil
}

func (ps structParsers) walk(p *Parser, expr ast.Expr, owner typeinfo.Type) (*Path, error) { // nolint:gocyclo
	expr = ast.Unparen(expr)

	if !owner.Deref().IsNamed() {
		return nil, codefmt.Errorf(p, expr, "field of anonymous struct is not supported")
	}

	// Unwrap convgen.Field~ function calls
	// e.g., convgen.FieldGetter(T{}.GetField) -> T{}.GetField
	if call, ok := expr.(*ast.CallExpr); ok {
		callee := typeutil.Callee(p.Pkg().TypesInfo, call)
		if callee != nil && IsConvgenImport(callee.Pkg().Path()) && strings.HasPrefix(callee.Name(), "Field") {
			expr = call.Args[0]
		}
	}

	switch x := expr.(type) {
	case *ast.CompositeLit:
		err := codefmt.Errorf(p, expr, "field should belong to %t{}; got %c", owner, expr)

		// Expression: T{}
		if len(x.Elts) != 0 {
			// T{...}
			return nil, err
		}

		id, ok := ast.Unparen(x.Type).(*ast.Ident)
		if !ok {
			// struct{}{}
			return nil, err
		}

		t := typeinfo.TypeOf(p.Pkg().TypesInfo.TypeOf(id))
		if !t.Identical(owner.Deref()) {
			// U{} where U != T
			return nil, err
		}

		obj := p.Pkg().TypesInfo.ObjectOf(id)
		return &Path{StructField: []types.Object{obj}, Pos: id.Pos()}, nil

	case *ast.CallExpr:
		err := codefmt.Errorf(p, expr, "field should belong to (*%t)(nil); got %c", owner, expr)

		// Expression: (*T)(nil)
		if len(x.Args) != 1 || !p.IsNil(x.Args[0]) {
			// (nil) expected but got (), (nonil), (...)
			return nil, err
		}

		fun := ast.Unparen(x.Fun)
		star, ok := fun.(*ast.StarExpr)
		if !ok {
			// f(nil) where f != *T
			return nil, err
		}

		id, ok := ast.Unparen(star.X).(*ast.Ident)
		if !ok {
			// (*struct{})(nil)
			return nil, err
		}

		t := typeinfo.TypeOf(p.Pkg().TypesInfo.TypeOf(id))
		if !t.Identical(owner.Deref()) {
			// (*U)(nil) where U != T
			return nil, err
		}

		obj := p.Pkg().TypesInfo.ObjectOf(id)
		return &Path{StructField: []types.Object{obj}, Pos: id.Pos()}, nil

	case *ast.UnaryExpr:
		// Expression: &T{}
		if x.Op != token.AND {
			return nil, codefmt.Errorf(p, codefmt.Pos(x.OpPos), "cannot use %s operation for field; got %c", x.Op, expr)
		}
		return ps.walk(p, x.X, owner)

	case *ast.SelectorExpr:
		// Expression: T{}.Field
		path, err := ps.walk(p, x.X, owner)
		if err != nil {
			return nil, err
		}

		obj := p.Pkg().TypesInfo.ObjectOf(x.Sel)
		if obj == nil {
			// Actually, impossible if type checking is done.
			return nil, codefmt.Errorf(p, expr, "cannot find field %q in %t", x.Sel.Name, owner)
		}

		path.StructField = append(path.StructField, obj)
		path.Pos = obj.Pos()
		return path, nil

	case *ast.StarExpr:
		// Expression: (*(T{}.Field))
		return ps.walk(p, x.X, owner)

	case *ast.IndexExpr, *ast.IndexListExpr:
		// Expression: ...[i]
		return nil, codefmt.Errorf(p, expr, "cannot use [] operation for field; got %c", expr)

	case *ast.BinaryExpr:
		// Expression: ... + ...
		return nil, codefmt.Errorf(p, expr, "cannot use %s operation for field; got %c", x.Op, expr)
	}

	return nil, codefmt.Errorf(p, expr, "invalid field expression; got %c", expr)
}

func (ps structParsers) ValidatePath(p *Parser, path Path, at token.Pos) error {
	return nil
}

func (structParsers) ParsePkgX(p *Parser, expr ast.Expr) (*types.Package, error) {
	panic("struct pkg parser must not be called")
}

func (structParsers) ParsePkgY(p *Parser, expr ast.Expr) (*types.Package, error) {
	panic("struct pkg parser must not be called")
}
