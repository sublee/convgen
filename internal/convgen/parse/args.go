package parse

import (
	"errors"
	"go/ast"
	"go/constant"
	"go/token"
	"strconv"

	"github.com/sublee/convgen/internal/codefmt"
)

type arg interface {
	bool | string
}

func parseArgExpr[T arg](p *Parser, expr ast.Expr) (T, error) {
	var v T
	switch any(v).(type) {
	case bool:
		tv := p.Pkg().TypesInfo.Types[expr]
		if tv.Value == nil || tv.Value.Kind() != constant.Bool {
			return v, codefmt.Errorf(p, expr, "%s is not bool literal", expr)
		}

		x := constant.BoolVal(tv.Value)
		v = any(x).(T)

	case string:
		lit, ok := expr.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return v, codefmt.Errorf(p, expr, "%s is not string literal", expr)
		}

		x, _ := strconv.Unquote(lit.Value)
		v = any(x).(T)

	default:
		panic("unreachable")
	}
	return v, nil
}

func parseArgs2[T1, T2 arg](p *Parser, call *ast.CallExpr) (T1, T2, error) {
	var v1 T1
	var v2 T2

	expr1, expr2, err := needArgs2(p, call)
	if err != nil {
		return v1, v2, err
	}

	var errs error

	v1, err = parseArgExpr[T1](p, expr1)
	errs = errors.Join(errs, err)

	v2, err = parseArgExpr[T2](p, expr2)
	errs = errors.Join(errs, err)

	return v1, v2, errs
}

func parseArgs3[T1, T2, T3 arg](p *Parser, call *ast.CallExpr) (T1, T2, T3, error) {
	var v1 T1
	var v2 T2
	var v3 T3

	expr1, expr2, expr3, err := needArgs3(p, call)
	if err != nil {
		return v1, v2, v3, err
	}

	var errs error

	v1, err = parseArgExpr[T1](p, expr1)
	errs = errors.Join(errs, err)

	v2, err = parseArgExpr[T2](p, expr2)
	errs = errors.Join(errs, err)

	v3, err = parseArgExpr[T3](p, expr3)
	errs = errors.Join(errs, err)

	return v1, v2, v3, errs
}

func parseArgs4[T1, T2, T3, T4 arg](p *Parser, call *ast.CallExpr) (T1, T2, T3, T4, error) {
	var v1 T1
	var v2 T2
	var v3 T3
	var v4 T4

	expr1, expr2, expr3, expr4, err := needArgs4(p, call)
	if err != nil {
		return v1, v2, v3, v4, err
	}

	var errs error

	v1, err = parseArgExpr[T1](p, expr1)
	errs = errors.Join(errs, err)

	v2, err = parseArgExpr[T2](p, expr2)
	errs = errors.Join(errs, err)

	v3, err = parseArgExpr[T3](p, expr3)
	errs = errors.Join(errs, err)

	v4, err = parseArgExpr[T4](p, expr4)
	errs = errors.Join(errs, err)

	return v1, v2, v3, v4, errs
}

func needArgs0(p *Parser, call *ast.CallExpr) error {
	if len(call.Args) != 0 {
		return codefmt.Errorf(p, call, "need no parameters")
	}
	return nil
}

func needArgs1(p *Parser, call *ast.CallExpr) (ast.Expr, error) {
	if len(call.Args) != 1 {
		return nil, codefmt.Errorf(p, call, "need 1 parameter")
	}
	return call.Args[0], nil
}

func needArgs2(p *Parser, call *ast.CallExpr) (ast.Expr, ast.Expr, error) {
	if len(call.Args) != 2 {
		return nil, nil, codefmt.Errorf(p, call, "need 2 parameters")
	}
	return call.Args[0], call.Args[1], nil
}

func needArgs3(p *Parser, call *ast.CallExpr) (ast.Expr, ast.Expr, ast.Expr, error) {
	if len(call.Args) != 3 {
		return nil, nil, nil, codefmt.Errorf(p, call, "need 3 parameters")
	}
	return call.Args[0], call.Args[1], call.Args[2], nil
}

func needArgs4(p *Parser, call *ast.CallExpr) (ast.Expr, ast.Expr, ast.Expr, ast.Expr, error) {
	if len(call.Args) != 4 {
		return nil, nil, nil, nil, codefmt.Errorf(p, call, "need 4 parameters")
	}
	return call.Args[0], call.Args[1], call.Args[2], call.Args[3], nil
}
