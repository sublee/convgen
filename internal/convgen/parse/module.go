package parse

import (
	"errors"
	"go/ast"
	"go/token"
	"iter"
	"slices"

	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/typeinfo"
)

// Module is a shared context for converters. It holds a registry of converters
// to make them available to each other. It also contains global configuration
// options. Every converters belong to a module inherit its configuration.
type Module struct {
	// Name is the module name user gave when declaring the module as a
	// variable. It can be empty if the module is declared inline.
	Name string

	// Config holds all configuration options defined in this module.
	Config Config

	// Lookup holds all converters and functions registered in this module.
	*typeinfo.Lookup[typeinfo.Func]

	// Configs and Lookups for specific kinds of converters.
	ConfigForStruct *Config
	ConfigForUnion  *Config
	ConfigForEnum   *Config
	LookupForStruct *typeinfo.Lookup[typeinfo.Func]
	LookupForUnion  *typeinfo.Lookup[typeinfo.Func]
	LookupForEnum   *typeinfo.Lookup[typeinfo.Func]
}

// ParseModules finds and parses all convgen.Module calls in the parsed files.
func (p *Parser) ParseModules() (map[token.Pos]*Module, error) {
	var errs error
	mods := make(map[token.Pos]*Module)

	for _, file := range p.ConvgenGoFiles() {
		for id, call := range p.FindModules(file) {
			name := id.Name
			if name == "_" {
				name = ""
			}

			mod, err := p.ParseModule(call, name)
			mods[id.Pos()] = mod
			errs = errors.Join(errs, err)
		}
	}

	return mods, errs
}

// FindModules collects and iterates package-level [convgen.Module] calls. It
// does not collect inline calls.
func (p *Parser) FindModules(file *ast.File) iter.Seq2[*ast.Ident, *ast.CallExpr] {
	return func(yield func(*ast.Ident, *ast.CallExpr) bool) {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range gen.Specs {
				val, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				for i, id := range val.Names {
					if len(val.Values) <= i {
						break
					}

					call, ok := ast.Unparen(val.Values[i]).(*ast.CallExpr)
					if !ok || !p.IsDirective(call, "Module") {
						continue
					}

					if !yield(id, call) {
						return
					}
				}
			}
		}
	}
}

// ParseModule parses a [convgen.Module] call expression and returns a new
// module.
func (p *Parser) ParseModule(call *ast.CallExpr, name string) (*Module, error) {
	// Chain of For* after NewModule
	calls := []*ast.CallExpr{call}
	for {
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			break
		}

		prevCall, ok := sel.X.(*ast.CallExpr)
		if !ok {
			break
		}

		calls = append(calls, prevCall)
		call = prevCall
	}
	slices.Reverse(calls)

	var cfg Config
	var errs error
	if err := p.ParseConfig(&cfg, calls[0].Args, nil); err != nil {
		errs = errors.Join(errs, err)
	}

	for _, call := range calls[1:] {
		switch call.Fun.(*ast.SelectorExpr).Sel.Name {
		case "ForStruct":
			cfg.ForStruct = &Config{}
			err := p.ParseConfig(cfg.ForStruct, call.Args, nil)
			errs = errors.Join(errs, err)
		case "ForUnion":
			cfg.ForUnion = &Config{}
			err := p.ParseConfig(cfg.ForUnion, call.Args, nil)
			errs = errors.Join(errs, err)
		case "ForEnum":
			cfg.ForEnum = &Config{}
			err := p.ParseConfig(cfg.ForEnum, call.Args, nil)
			errs = errors.Join(errs, err)
		default:
			panic("unexpected module chain")
		}
	}

	// Register imported functions
	lookup, err := p.newModuleLookup(cfg, nil)
	errs = errors.Join(errs, err)

	if cfg.ForStruct != nil {
		_, err = p.newModuleLookup(*cfg.ForStruct, lookup)
		errs = errors.Join(errs, err)
	}
	if cfg.ForUnion != nil {
		_, err = p.newModuleLookup(*cfg.ForUnion, lookup)
		errs = errors.Join(errs, err)
	}
	if cfg.ForEnum != nil {
		_, err = p.newModuleLookup(*cfg.ForEnum, lookup)
		errs = errors.Join(errs, err)
	}

	return &Module{Name: name, Config: cfg, Lookup: lookup}, errs
}

func (p *Parser) newModuleLookup(cfg Config, old *typeinfo.Lookup[typeinfo.Func]) (*typeinfo.Lookup[typeinfo.Func], error) {
	var errs error

	lookup := typeinfo.NewLookup[typeinfo.Func]()
	if old != nil {
		for fn := range old.Range() {
			lookup.Put(fn)
		}
	}

	for i, fn := range cfg.Funcs {
		if oldFn, ok := lookup.Put(fn); !ok {
			err := codefmt.Errorf(p, cfg.FuncExprs[i], `duplicate %t to %t converter
	previous import of %o at %b`,
				fn.X(), fn.Y(),
				oldFn, oldFn)
			errs = errors.Join(errs, err)
		}
	}

	return lookup, errs
}

// ParseModuleArg parses a Convgen module type argument from the given
// expression.
func (p *Parser) ParseModuleArg(expr ast.Expr, mods map[token.Pos]*Module) (*Module, error) {
	expr = ast.Unparen(expr)

	// Inline Module Declaration
	// =========================
	//
	//	var conv = convgen.Struct[X, Y](convgen.Module(...))
	//	                                ^^^^^^^^^^^^^^^^^^^
	// This type of module is anonymous and cannot be referred by other
	// converters. But it is still useful when the converter requires another
	// implicit converters. The implicit converters will inherit the module's
	// configuration.
	if call, ok := expr.(*ast.CallExpr); ok && p.IsDirective(call, "Module") {
		return p.ParseModule(call, "")
	}

	// Validate identifier
	id, ok := expr.(*ast.Ident)
	if !ok {
		return nil, codefmt.Errorf(p, expr, "module must be convgen.Module() or package-level variable")
	}

	t := typeinfo.TypeOf(p.Pkg().TypesInfo.TypeOf(id))

	// Nil Module
	// ==========
	//
	//	var conv = convgen.Struct[X, Y](nil)
	//	                                ^^^
	// Nil indicates a new empty module with no configuration.
	if t.IsNil() {
		// The module is nil which is legal. It means a single converter has its
		// own module.
		return NilModule(), nil
	}

	// Package-level Module
	// ====================
	//
	//	var (
	//		mod = convgen.Module(...)
	//		^^^
	//		x2y = convgen.Struct[X, Y](mod)
	//		a2b = convgen.Struct[A, B](mod)
	//	)
	//
	// This is the most common way to declare and use a module. Multiple
	// converters can belong to the same package-level module.
	modPos := p.Pkg().TypesInfo.ObjectOf(id).Pos()
	mod, ok := mods[modPos]
	if !ok {
		return nil, codefmt.Errorf(p, expr, "cannot find %q module declared by convgen.Module", id.Name)
	}
	return mod, nil
}

// NilModule returns a new empty module with no configuration.
func NilModule() *Module {
	return &Module{Lookup: typeinfo.NewLookup[typeinfo.Func]()}
}
