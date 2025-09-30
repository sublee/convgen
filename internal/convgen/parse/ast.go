package parse

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strconv"
)

// evalBoolLit evaluates a bool expression. Returns (b, ok) where b is the
// evaluated bool value. Since "true" and "false" is a constant, we need
// [types.Info] to guess the value.
func evalBoolLit(expr ast.Expr, info *types.Info) (bool, bool) {
	tv := info.Types[expr]
	if tv.Value == nil || tv.Value.Kind() != constant.Bool {
		return false, false
	}
	return constant.BoolVal(tv.Value), true
}

// evalStringLit evaluates a string expression. Returns (s, ok) where s is the
// evaluated string value.
func evalStringLit(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	s, _ := strconv.Unquote(lit.Value)
	return s, true
}

// tailIdent extracts the rightmost [ast.Ident] from the expression.
//
//	Foo{}
//	^^^
//	Foo{}.Bar
//	      ^^^
//	(*Foo)(nil).Bar.Baz
//	                ^^^
func tailIdent(expr ast.Expr) (*ast.Ident, bool) {
	expr = ast.Unparen(expr)
	switch expr := expr.(type) {
	case *ast.Ident:
		// foo
		// ^^^
		return expr, true
	case *ast.SelectorExpr:
		// foo.bar.baz
		//         ^^^
		return tailIdent(expr.Sel)
	}
	return nil, false
}
