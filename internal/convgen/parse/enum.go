package parse

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/typeinfo"
)

// ParseEnumMember parses an expression of an enum member of the given enum
// type.  If the type of the expression is the given enum type, it returns the
// name of the value and the package where the type is defined.
func (p *Parser) ParseEnumMember(expr ast.Expr, enum typeinfo.Type) (*types.Const, error) {
	expr = ast.Unparen(expr)

	if !enum.IsNamed() || !enum.IsBasic() {
		panic(codefmt.Sprintf(p, "not enum: %t", enum))
	}

	// Unwrap package
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		id, ok := sel.X.(*ast.Ident)
		if !ok {
			return nil, codefmt.Errorf(p, expr, "enum member must be package-level constant; got %c", expr)
		}

		pkgName := p.Pkg().TypesInfo.ObjectOf(id)
		if _, ok := pkgName.(*types.PkgName); !ok {
			return nil, codefmt.Errorf(p, expr, "enum member must be package-level constant; got %c", expr)
		}

		expr = sel.Sel
	}

	id, ok := expr.(*ast.Ident)
	if !ok {
		return nil, codefmt.Errorf(p, expr, "enum member must be package-level constant; got %c", expr)
	}

	obj := p.Pkg().TypesInfo.ObjectOf(id)
	if !types.Identical(obj.Type(), enum.Type()) {
		return nil, codefmt.Errorf(p, expr, "%c is not member of enum %t; belongs to %t", expr, enum, obj.Type())
	}

	con, ok := obj.(*types.Const)
	if !ok {
		return nil, codefmt.Errorf(p, expr, "enum member must be package-level constant; got %c", expr)
	}

	return con, nil
}

type enumParsers struct{ x, y typeinfo.Type }

func (ps enumParsers) parse(p *Parser, expr ast.Expr, enum typeinfo.Type) (*Path, *types.Package, error) {
	con, err := p.ParseEnumMember(expr, enum)
	if err != nil {
		return nil, nil, err
	}
	return &Path{EnumMember: con, Pos: con.Pos()}, con.Pkg(), nil
}

func (ps enumParsers) ParsePathX(p *Parser, expr ast.Expr) (*Path, error) {
	path, _, err := ps.parse(p, expr, ps.x)
	return path, err
}

func (ps enumParsers) ParsePathY(p *Parser, expr ast.Expr) (*Path, error) {
	path, _, err := ps.parse(p, expr, ps.y)
	return path, err
}

func (ps enumParsers) ValidatePath(p *Parser, path Path, at token.Pos) error {
	return nil
}

func (ps enumParsers) ParsePkgX(p *Parser, expr ast.Expr) (*types.Package, error) {
	_, pkg, err := ps.parse(p, expr, ps.x)
	return pkg, err
}

func (ps enumParsers) ParsePkgY(p *Parser, expr ast.Expr) (*types.Package, error) {
	_, pkg, err := ps.parse(p, expr, ps.y)
	return pkg, err
}
