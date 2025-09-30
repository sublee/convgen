package parse

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/typeinfo"
)

type unionParsers struct {
	x, y     typeinfo.Type
	parsed   map[token.Pos]Path
	parsedAt map[token.Pos]token.Pos
}

func newUnionParsers(x, y typeinfo.Type) unionParsers {
	return unionParsers{
		x:        x,
		y:        y,
		parsed:   make(map[token.Pos]Path),
		parsedAt: make(map[token.Pos]token.Pos),
	}
}

func (ps unionParsers) parse(p *Parser, expr ast.Expr, union typeinfo.Type) (*Path, *types.Package, error) {
	if !union.IsInterface() {
		panic("not interface")
	}

	t := typeinfo.TypeOf(p.Pkg().TypesInfo.TypeOf(expr))

	if t.IsNamed() {
		if types.Implements(t.Named, union.Interface) {
			return &Path{UnionImpl: t.Named, Pos: t.Named.Obj().Pos()}, t.Pkg(), nil
		}
	}

	if t.IsPointer() && t.Elem.IsNamed() {
		if types.Implements(t.Pointer, union.Interface) {
			return &Path{UnionImpl: t.Pointer, Pos: t.Elem.Named.Obj().Pos()}, t.Pkg(), nil
		}
	}

	return nil, nil, codefmt.Errorf(p, expr, "%t does not implement %t", t, union)
}

func (ps unionParsers) ParsePathX(p *Parser, expr ast.Expr) (*Path, error) {
	path, _, err := ps.parse(p, expr, ps.x)
	return path, err
}

func (ps unionParsers) ParsePathY(p *Parser, expr ast.Expr) (*Path, error) {
	path, _, err := ps.parse(p, expr, ps.y)
	return path, err
}

func (ps unionParsers) ValidatePath(p *Parser, path Path, at token.Pos) error {
	if oldPath, ok := ps.parsed[path.Pos]; ok {
		oldAt := ps.parsedAt[path.Pos]

		new := typeinfo.TypeOf(path.UnionImpl)
		old := typeinfo.TypeOf(oldPath.UnionImpl)

		if new.IsPointer() && !old.IsPointer() {
			return codefmt.Errorf(p, codefmt.Pos(at), `%t redeclared as pointer type %t
	previous declaration at %b`, old, new, oldAt)
		}

		if !new.IsPointer() && old.IsPointer() {
			return codefmt.Errorf(p, codefmt.Pos(at), `%t redeclared as non-pointer type %t
	previous declaration at %b`, old, new, oldAt)
		}
	}
	ps.parsed[path.Pos] = path
	ps.parsedAt[path.Pos] = at
	return nil
}

func (ps unionParsers) ParsePkgX(p *Parser, expr ast.Expr) (*types.Package, error) {
	_, pkg, err := ps.parse(p, expr, ps.x)
	return pkg, err
}

func (ps unionParsers) ParsePkgY(p *Parser, expr ast.Expr) (*types.Package, error) {
	_, pkg, err := ps.parse(p, expr, ps.y)
	return pkg, err
}
