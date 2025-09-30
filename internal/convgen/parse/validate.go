package parse

import (
	"errors"
	"go/ast"
	"strings"

	"github.com/sublee/convgen/internal/codefmt"
)

// Validate checks for usages outside expected paths. It collects all errors
// instead of stopping at the first error.
//
// Many validation rules are implemented in the expected paths by narrow parsing
// functions. But some rules need to be checked globally. That's what this
// function does.
func (p *Parser) Validate() error {
	var errs error
	for _, file := range p.Pkg().Syntax {
		errs = errors.Join(errs, p.validateConstraint(file))
		errs = errors.Join(errs, p.validateModulesInsideFunc(file))
		errs = errors.Join(errs, p.validateAssignedDirectives(file))
	}
	return errs
}

// validateConstraint checks if files importing "github.com/sublee/convgen" have
// "//go:build convgen" constraint.
func (p *Parser) validateConstraint(file *ast.File) error {
	// Find convgen import
	var convgenImport *ast.ImportSpec
	for _, imp := range file.Imports {
		if IsConvgenImport(strings.Trim(imp.Path.Value, `"`)) {
			convgenImport = imp
			break
		}
	}
	if convgenImport == nil {
		return nil // No convgen import found
	}

	// Check for "//go:build convgen" constraint
	if hasGoBuildConvgen(file) {
		return nil // Constraint satisfied
	}

	// This file imports convgen but has no "//go:build convgen" constraint
	return codefmt.Errorf(p, convgenImport, `file must have "//go:build convgen" constraint when importing convgen`)
}

// validateModuleInsideFunc validates that no module is assigned inside a
// function.
//
// Convgen module must be determined at build time and must not give the
// impression that anything can be modified at runtime. Therefore, assignments
// to package-level variables or direct inlining into injectors are allowed, but
// assignments inside functions are not allowed. If we allow assignments inside
// functions, it could give the false expectation that reassigning the variable
// would override the module's configuration.
//
//	// good: package-level variable
//	var mod = convgen.Module(...)
//
//	// good: inline in injector
//	var conv = convgen.Struct[X, Y](convgen.Module(...))
//
//	// bad: inside function
//	func MyFunc() {
//		mod := convgen.Module(...)
//		...
//	}
func (p *Parser) validateModulesInsideFunc(file *ast.File) error {
	if !hasGoBuildConvgen(file) {
		return nil
	}

	var errs error
	validate := func(fun ast.Node) {
		ast.Inspect(fun, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.ValueSpec, *ast.AssignStmt:
				ast.Inspect(node, func(node ast.Node) bool {
					call, ok := node.(*ast.CallExpr)
					if !ok || !p.IsDirective(call, "Module") {
						return true
					}

					err := codefmt.Errorf(p, call, "cannot assign module to variable inside function")
					errs = errors.Join(errs, err)
					return false
				})
				return false
			}
			return true
		})
	}

	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			// func F() { ... }
			validate(decl)

		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				for _, value := range vs.Values {
					if fun, ok := value.(*ast.FuncLit); ok {
						// var F = func() { ... }
						validate(fun)
					}
				}
			}
		}
	}

	return errs
}

// validateAssignedDirectives checks illegal assignments of Convgen directives.
//
// Only module and injectors are allowed to be assigned to variables. Other
// directives, for example options, cannot be assigned. This is to prevent
// remaining Convgen import after code generation.
func (p *Parser) validateAssignedDirectives(file *ast.File) error {
	if !hasGoBuildConvgen(file) {
		return nil
	}

	var errs error
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.ValueSpec, *ast.AssignStmt:
			ast.Inspect(node, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				directive, ok := p.GetDirective(call)
				if !ok {
					return true
				}

				// Module and injectors cannot be assigned to variables.
				switch directive {
				case "Module":
					return false
				case "Struct", "StructErr":
					return false
				case "Union", "UnionErr":
					return false
				case "Enum", "EnumErr":
					return false
				}

				// Other directives are allowed to be assigned.
				err := codefmt.Errorf(p, call, "cannot assign %s to variable", directive)
				errs = errors.Join(errs, err)
				return false
			})
			return false
		}
		return true
	})
	return errs
}
