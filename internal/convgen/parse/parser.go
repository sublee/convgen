package parse

import (
	"errors"
	"fmt"
	"go/ast"
	"go/build/constraint"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/typeinfo"
)

func IsConvgenImport(path string) bool {
	// Source code from "wire/internal/wire/parse.go".
	const vendorPart = "vendor/"
	if i := strings.LastIndex(path, vendorPart); i != -1 && (i == 0 || path[i-1] == '/') {
		path = path[i+len(vendorPart):]
	}
	return path == "github.com/sublee/convgen"
}

// Parser parses an AST of the underlying package to collect convgen converters.
type Parser struct{ pkg *packages.Package }

func (p *Parser) Pkg() *packages.Package { return p.pkg }

// New creates a new [Parser].
func New(pkg *packages.Package) (*Parser, error) {
	if pkg.Name == "" {
		return nil, fmt.Errorf("need pkg name")
	}
	if pkg.PkgPath == "" {
		return nil, fmt.Errorf("need pkg path")
	}
	if pkg.Types == nil {
		return nil, fmt.Errorf("need pkg types")
	}
	if pkg.Fset == nil {
		return nil, fmt.Errorf("need pkg fset")
	}
	if pkg.Syntax == nil {
		return nil, fmt.Errorf("need pkg syntax")
	}
	if pkg.TypesInfo == nil {
		return nil, fmt.Errorf("need pkg types info")
	}
	return &Parser{pkg: pkg}, nil
}

func (p *Parser) IsNil(expr ast.Expr) bool {
	expr = ast.Unparen(expr)

	// nil
	if id, ok := expr.(*ast.Ident); ok {
		if id.Name == "nil" {
			return true
		}
	}

	// T(nil)
	if call, ok := expr.(*ast.CallExpr); ok {
		fun := ast.Unparen(call.Fun)
		if !call.Ellipsis.IsValid() && len(call.Args) == 1 {
			switch fun.(type) {
			case *ast.ArrayType, *ast.StructType, *ast.FuncType, *ast.InterfaceType, *ast.MapType, *ast.ChanType:
				return p.IsNil(call.Args[0])
			}
		}
	}

	return false
}

// ParseFunc parses a function expression. If hasErr is true, the function must
// return an error as the last return value. The function is used for
// [convgen.ImportFunc], [convgen.MatchFunc], and their Err variants.
func (p *Parser) ParseFunc(expr ast.Expr, hasErr bool) (typeinfo.Func, error) {
	expr = ast.Unparen(expr)

	var fn typeinfo.Func
	if lit, ok := expr.(*ast.FuncLit); ok {
		fn_, err := typeinfo.FuncLitOf[typeinfo.BothXY](p.pkg, lit)
		if err != nil {
			return nil, codefmt.Errorf(p, expr, "%s", err.Error())
		}
		fn = fn_
	} else {
		if p.IsNil(expr) {
			return nil, codefmt.Errorf(p, expr, "cannot use nil as function")
		}

		id, ok := tailIdent(expr)
		if !ok {
			return nil, codefmt.Errorf(p, expr, "cannot use %c as function", expr)
		}

		obj, err := p.getFuncObj(id, expr)
		if err != nil {
			return nil, codefmt.Errorf(p, expr, "%s", err.Error())
		}
		fn_, err := typeinfo.FuncOf[typeinfo.BothXY](obj)
		if err != nil {
			return nil, codefmt.Errorf(p, expr, "%s", err.Error())
		}
		fn = fn_
	}

	if hasErr && !fn.HasErr() {
		return nil, codefmt.Errorf(p, expr, "function must return error") // unreachable
	} else if !hasErr && fn.HasErr() {
		return nil, codefmt.Errorf(p, expr, "function must not return error") // unreachable
	}
	return fn, nil
}

// ParseErrWrap parses an error wrapper function expression. The function must
// take an error as the only parameter and return an error. The function is used
// for [convgen.ErrWrap].
func (p *Parser) ParseErrWrap(expr ast.Expr) (typeinfo.Func, error) {
	expr = ast.Unparen(expr)

	var fn typeinfo.Func
	if lit, ok := expr.(*ast.FuncLit); ok {
		fn_, err := typeinfo.FuncLitOf[typeinfo.OnlyX](p.pkg, lit)
		if err != nil {
			return nil, codefmt.Errorf(p, expr, "%s", err.Error())
		}
		fn = fn_
	} else {
		id, ok := tailIdent(expr)
		if !ok {
			return nil, codefmt.Errorf(p, expr, "cannot use %c as error wrapper", expr)
		}

		obj, err := p.getFuncObj(id, expr)
		if err != nil {
			if errors.Is(err, errNilFuncObj) {
				return nil, codefmt.Errorf(p, expr, "cannot use nil as error wrapper")
			}
			return nil, codefmt.Errorf(p, expr, "%s", err.Error())
		}

		fn_, err := typeinfo.FuncOf[typeinfo.OnlyX](obj)
		if err != nil {
			return nil, codefmt.Errorf(p, expr, "%s", err.Error())
		}
		fn = fn_
	}

	if !fn.X().IsError() {
		return nil, codefmt.Errorf(p, expr, "error wrapper parameter must be error") // unreachable
	}
	if !fn.HasErr() {
		return nil, codefmt.Errorf(p, expr, "error wrapper must return error") // unreachable
	}
	return fn, nil
}

// GetDirective returns the name of the Convgen directive function if the call
// expression is a Convgen directive. Otherwise, it returns false.
func (p *Parser) GetDirective(call *ast.CallExpr) (string, bool) {
	callee := typeutil.Callee(p.Pkg().TypesInfo, call)
	if callee == nil {
		return "", false
	}

	pkg := callee.Pkg()
	if pkg == nil {
		// Built-in functions like panic()
		return "", false
	}

	if !IsConvgenImport(pkg.Path()) {
		// Not Convgen function
		return "", false
	}

	return callee.Name(), true
}

// IsDirective checks if the call expression is a Convgen directive with the
// given name. If name is empty, it checks if the call is any Convgen directive.
func (p *Parser) IsDirective(call *ast.CallExpr, name string) bool {
	calleeName, ok := p.GetDirective(call)
	if !ok {
		return false
	}

	if name == "" {
		// Any convgen directive
		return true
	}

	return calleeName == name
}

// ConvgenGoFiles returns the Go files that have a "//go:build convgen"
// constraint.
func (p *Parser) ConvgenGoFiles() []*ast.File {
	var files []*ast.File
	for _, file := range p.Pkg().Syntax {
		if hasGoBuildConvgen(file) {
			files = append(files, file)
		}
	}
	return files
}

var errNilFuncObj = errors.New("nil function object")

// getFuncObj returns the object of the function in the expression. If the expression is a generic function instance, it creates an object with a generic name,
// but a function literal signature
func (p *Parser) getFuncObj(id *ast.Ident, expr ast.Expr) (types.Object, error) {
	obj := p.Pkg().TypesInfo.ObjectOf(id)
	if _, ok := obj.(*types.Nil); ok {
		return nil, errNilFuncObj
	}
	objType := obj.Type()
	declarationSignature, ok := objType.(*types.Signature) // This holds the generic function declaration (func[T any](T) string)
	if !ok {
		return nil, fmt.Errorf("cannot get object signature of type %s", objType.String())
	}
	instanceType := p.Pkg().TypesInfo.TypeOf(expr)
	instanceSignature, ok := instanceType.(*types.Signature) // This holds the instance (func (int) string)
	if !ok {
		return nil, fmt.Errorf("cannot get object signature of type %s", instanceType.String())
	}
	if declarationSignature.TypeParams().Len() > 0 {
		var name strings.Builder
		name.WriteString(obj.Name())
		switch e := expr.(type) {
		case *ast.IndexExpr:
			name.WriteString("[")
			name.WriteString(codefmt.New(p.pkg).Expr(e.Index))
			name.WriteString("]")
		case *ast.IndexListExpr:
			name.WriteString("[")
			for i, idx := range e.Indices {
				if i > 0 {
					name.WriteString(", ")
				}
				name.WriteString(codefmt.New(p.pkg).Expr(idx))
			}
			name.WriteString("]")
		}
		return types.NewFunc(obj.Pos(), obj.Pkg(), name.String(), instanceSignature), nil
	}
	return obj, nil
}

// hasGoBuildConvgen checks if the file has a "//go:build convgen" constraint.
func hasGoBuildConvgen(file *ast.File) bool {
	ok := false
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if constraint.IsGoBuild(comment.Text) {
				expr, _ := constraint.Parse(comment.Text)
				expr.Eval(func(tag string) bool {
					if tag == "convgen" {
						ok = true
					}
					return true
				})
			}
		}
	}
	return ok
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
	case *ast.IndexExpr:
		// foo[T]
		// ^^^
		return tailIdent(expr.X)
	case *ast.IndexListExpr:
		// foo[T, U]
		// ^^^
		return tailIdent(expr.X)
	}
	return nil, false
}
