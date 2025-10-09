package parse

import (
	"errors"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/sublee/convgen/internal/codefmt"
)

// Validate checks for usages outside expected paths. It collects all errors
// instead of stopping at the first error.
//
// Many validation rules are implemented in the expected paths by narrow parsing
// functions. But some rules need to be checked globally. That's what this
// function does.
func (p *Parser) Validate(mods map[token.Pos]*Module) error {
	var errs error
	for _, file := range p.Pkg().Syntax {
		errs = errors.Join(errs, p.validateConstraint(file))
		errs = errors.Join(errs, p.validateAssignedDirectives(file))
	}
	errs = errors.Join(errs, p.validateModuleUsages(mods))
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

// validateModuleUsages checks illegal references to modules.
//
// Modules are only allowed to be assigned to variables (except exported ones)
// or used as arguments to Convgen directives. Any other usages are illegal,
// because modules will be removed at code generation, and any remaining
// references to modules will cause compilation errors.
func (p *Parser) validateModuleUsages(mods map[token.Pos]*Module) error {
	var errs error
	blanks := p.findBlankValues()
	for _, file := range p.Pkg().Syntax {
		astutil.Apply(file, func(c *astutil.Cursor) bool {
			if call, ok := c.Node().(*ast.CallExpr); ok {
				if p.IsDirective(call, "") {
					// A module can be used by Convgen directives. That's fine.
					return false
				}
				return true
			}

			// We will check all use of identifiers.
			id, ok := c.Node().(*ast.Ident)
			if !ok {
				return true
			}

			if _, ok := blanks[id.Pos()]; ok {
				// Assigned to blank identifier. That's fine.
				return false
			}

			obj := p.pkg.TypesInfo.ObjectOf(id)
			if obj == nil {
				// Cannot resolve identifier. Skip it.
				return false
			}

			mod, ok := mods[obj.Pos()]
			if !ok {
				// Not a module identifier. Skip it.
				return false
			}

			if id.IsExported() {
				err := codefmt.Errorf(p, id, "cannot export module %q; removed at code generation", mod.Name)
				errs = errors.Join(errs, err)
				return false
			}

			if id.Pos() == obj.Pos() {
				// This is the module identifier declaration. That's fine.
				return false
			}

			err := codefmt.Errorf(p, id, "cannot use module %q outside convgen directives; removed at code generation", mod.Name)
			errs = errors.Join(errs, err)
			return false
		}, nil)
	}
	return errs
}

// findBlankValues finds all expressions assigned to blank identifier (_) by its
// position.
func (p *Parser) findBlankValues() map[token.Pos]struct{} {
	blanks := make(map[token.Pos]struct{})
	for _, file := range p.Pkg().Syntax {
		astutil.Apply(file, func(c *astutil.Cursor) bool {
			switch node := c.Node().(type) {
			case *ast.ValueSpec:
				if len(node.Names) == len(node.Values) {
					// var a = b
					// var c, d = e, f
					for i, name := range node.Names {
						if name.Name == "_" {
							blanks[node.Values[i].Pos()] = struct{}{}
						}
					}
				} else if len(node.Names) > 1 && len(node.Values) == 1 {
					// var a, b = f()
					for _, name := range node.Names {
						if name.Name == "_" {
							blanks[node.Values[0].Pos()] = struct{}{}
						}
					}
				}
			case *ast.AssignStmt:
				if len(node.Lhs) == len(node.Rhs) {
					// a := b
					// c, d := e, f
					for i, lh := range node.Lhs {
						if id, ok := lh.(*ast.Ident); ok && id.Name == "_" {
							blanks[node.Rhs[i].Pos()] = struct{}{}
						}
					}
				} else if len(node.Lhs) > 1 && len(node.Rhs) == 1 {
					// a, b := f()
					for _, lh := range node.Lhs {
						if id, ok := lh.(*ast.Ident); ok && id.Name == "_" {
							blanks[node.Rhs[0].Pos()] = struct{}{}
						}
					}
				}
			}
			return true
		}, nil)
	}
	return blanks
}
