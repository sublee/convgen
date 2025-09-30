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

func (ps structParsers) parse(p *Parser, expr ast.Expr, owner typeinfo.Type) (*Path, error) {
	var parse func(ast.Expr, typeinfo.Type) (*Path, error)
	parse = func(expr ast.Expr, owner typeinfo.Type) (*Path, error) {
		expr = ast.Unparen(expr)

		// Unwrap convgen.Field~ function calls
		// e.g., convgen.FieldGetter(T{}.GetField) -> T{}.GetField
		if call, ok := expr.(*ast.CallExpr); ok {
			callee := typeutil.Callee(p.Pkg().TypesInfo, call)
			if callee != nil && IsConvgenImport(callee.Pkg().Path()) && strings.HasPrefix(callee.Name(), "Field") {
				expr = call.Args[0]
			}
		}

		fieldErr := codefmt.Errorf(p, expr, "field must belong to %t{}; got %c", owner, expr)
		methodErr := codefmt.Errorf(p, expr, "method must belong to %t{}, &%t{}, or (*%t)(nil); got %c", owner, owner, owner, expr)

		switch x := expr.(type) {
		case *ast.CompositeLit:
			// Expression: T{}
			if len(x.Elts) != 0 {
				// T{...}
				return nil, fieldErr
			}

			t := typeinfo.TypeOf(p.Pkg().TypesInfo.TypeOf(x.Type))
			if !t.Identical(owner.Deref()) {
				// U{}
				return nil, fieldErr
			}

			id, ok := ast.Unparen(x.Type).(*ast.Ident)
			if !ok {
				// struct{}{}
				return nil, fieldErr
			}

			obj := p.Pkg().TypesInfo.ObjectOf(id)
			return &Path{StructField: []types.Object{obj}, Pos: id.Pos()}, nil

		case *ast.UnaryExpr:
			// Expression: &T{}
			if x.Op != token.AND {
				return nil, methodErr
			}

			comp, ok := x.X.(*ast.CompositeLit)
			if !ok {
				return nil, methodErr
			}

			path, err := parse(comp, owner)
			if err != nil {
				return nil, methodErr
			}
			return path, nil

		case *ast.CallExpr:
			// Expression: (*T)(nil)
			if len(x.Args) != 1 || !p.IsNil(x.Args[0]) {
				return nil, methodErr
			}

			fun := ast.Unparen(x.Fun)
			star, ok := fun.(*ast.StarExpr)
			if !ok {
				return nil, methodErr
			}

			id, ok := ast.Unparen(star.X).(*ast.Ident)
			if !ok {
				return nil, methodErr
			}

			t := typeinfo.TypeOf(p.Pkg().TypesInfo.TypeOf(id))
			if !t.Identical(owner.Deref()) {
				return nil, methodErr
			}

			obj := p.Pkg().TypesInfo.ObjectOf(id)
			return &Path{StructField: []types.Object{obj}, Pos: id.Pos()}, nil

		case *ast.SelectorExpr:
			// Expression: T{}.Field
			path, err := parse(x.X, owner)
			if err != nil {
				return nil, err
			}

			obj := p.Pkg().TypesInfo.ObjectOf(x.Sel)
			path.StructField = append(path.StructField, obj)
			path.Pos = obj.Pos()
			return path, nil
		}

		return nil, fieldErr
	}
	return parse(expr, owner)
}

func (ps structParsers) ParsePathX(p *Parser, expr ast.Expr) (*Path, error) {
	return ps.parse(p, expr, ps.x)
}

func (ps structParsers) ParsePathY(p *Parser, expr ast.Expr) (*Path, error) {
	return ps.parse(p, expr, ps.y)
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
