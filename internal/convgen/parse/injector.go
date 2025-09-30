package parse

import (
	"errors"
	"go/ast"
	"go/token"
	"go/types"
	"iter"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/typeinfo"
)

// Injector represents a Convgen converter specification defined by an injector
// like convgen.Struct, convgen.Union, or convgen.Enum.
type Injector struct {
	typeinfo.Func
	Module *Module
	Config Config

	Struct      bool
	Union       bool
	Enum        bool
	EnumUnknown *types.Const

	pkg *packages.Package
	pos token.Pos

	Doc     *ast.CommentGroup
	Comment *ast.CommentGroup

	parent *Injector
}

// Pkg returns the package where the injector is called. Injector implements
// [codefmt.Pkger] by this method.
func (inj Injector) Pkg() *packages.Package { return inj.pkg }

// Pos returns the token position where the injector is called. Injector
// implements [codefmt.Poser] by this method.
func (inj Injector) Pos() token.Pos { return inj.pos }

// String returns a string representation of the injector. For example,
// "convgen.Struct[Foo, Bar]".
func (inj Injector) String() string {
	return inj.StringWithHasErr(inj.HasErr())
}

func (inj Injector) StringWithHasErr(hasErr bool) string {
	var buf strings.Builder
	switch {
	case inj.Struct:
		buf.WriteString("convgen.Struct")
	case inj.Union:
		buf.WriteString("convgen.Union")
	case inj.Enum:
		buf.WriteString("convgen.Enum")
	}
	if hasErr {
		buf.WriteString("Err")
	}
	codefmt.Fprintf(inj, &buf, "[%t, %t]", inj.X(), inj.Y())
	return buf.String()
}

// Fork copies the injector with replaced input and output types.
func (inj Injector) Fork(x, y typeinfo.Type) Injector {
	return Injector{
		Func: typeinfo.NewFunc(inj.Pkg().Types, "", x, y, inj.HasErr(), inj.HasOut()),

		Module: inj.Module,
		Config: inj.Config.Fork(),

		Struct: inj.Struct,
		Union:  inj.Union,
		Enum:   inj.Enum,

		pkg: inj.pkg,
		pos: inj.pos,

		parent: &inj,
	}
}

// Root returns the root injector in the hierarchy. The root injector is
// explicitly injected by user code. Otherwise, implicitly forked for a
// subconverter.
func (inj Injector) Root() Injector {
	if inj.parent == nil {
		return inj
	}
	return inj.parent.Root()
}

// ParseInjectors parses all [Injector]s from the AST.
func (p *Parser) ParseInjectors(ns codefmt.NS, mods map[token.Pos]*Module) ([]Injector, error) {
	var errs error
	var injs []Injector

	for _, file := range p.ConvgenGoFiles() {
		for inj, err := range p.parseInjectorsInFile(file, mods) {
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}
			injs = append(injs, inj)
		}
	}

	if errs != nil {
		return nil, errs
	}
	return injs, nil
}

// parseInjectorsInFile parses and yields [Injector]s in the given file.
func (p *Parser) parseInjectorsInFile(file *ast.File, mods map[token.Pos]*Module) iter.Seq2[Injector, error] {
	return func(yield func(Injector, error) bool) {
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

				if len(val.Names) != len(val.Values) {
					// Injectors should return exactly one value. The
					// assignment like this is invalid:
					// a, b := convgen.Struct[Foo, Bar](nil)
					continue
				}

				for i := range val.Values {
					call, ok := val.Values[i].(*ast.CallExpr)
					if !ok {
						continue
					}

					if !p.isInjector(call) {
						continue
					}

					id := val.Names[i]
					inj, err := p.parseInjector(id, call, val.Doc, val.Comment, mods)
					if err != nil {
						if !yield(Injector{}, err) {
							return
						}
					}

					if !yield(inj, nil) {
						return
					}
				}
			}
		}
	}
}

// isInjector checks if the given call expression is a converter injector call,
// such as convgen.Struct, convgen.Union, or convgen.Enum.
func (p *Parser) isInjector(call *ast.CallExpr) bool {
	if call == nil {
		return false
	}

	callee := typeutil.Callee(p.Pkg().TypesInfo, call)
	if callee == nil || callee.Pkg() == nil {
		return false
	}

	if !IsConvgenImport(callee.Pkg().Path()) {
		return false
	}

	switch callee.Name() {
	case "Struct", "StructErr":
		return true
	case "Union", "UnionErr":
		return true
	case "Enum", "EnumErr":
		return true
	}
	return false
}

// parseInjector parses an [Injector] from the given AST nodes.
func (p *Parser) parseInjector(id *ast.Ident, call *ast.CallExpr, doc, comment *ast.CommentGroup, mods map[token.Pos]*Module) (Injector, error) {
	var errs error
	inj := Injector{
		pkg:     p.Pkg(),
		pos:     call.Pos(),
		Doc:     doc,
		Comment: comment,
	}

	if id != nil && id.Name == "_" {
		return Injector{}, codefmt.Errorf(p, id, "cannot assign converter to blank identifier")
	}

	fn, err := typeinfo.FuncOf[typeinfo.BothXY](p.pkg.TypesInfo.ObjectOf(id))
	if err != nil {
		panic(err)
	}

	inj.Func = fn
	errs = errors.Join(errs, err)

	mod, err := p.ParseModuleArg(call.Args[0], mods)
	if err != nil {
		mod = NilModule() // Prevent nil panic to collect as many errors as possible
	}
	inj.Module = mod
	errs = errors.Join(errs, err)

	if errs != nil {
		return Injector{}, errs
	}

	var cfg Config
	var parsers parsers
	var opts []ast.Expr

	callee := typeutil.Callee(p.Pkg().TypesInfo, call)
	switch callee.Name() {
	case "Struct", "StructErr":
		inj.Struct = true
		cfg = mod.Config.ForkForStruct()
		parsers = structParsers{inj.X(), inj.Y()}
		opts = call.Args[1:]

	case "Union", "UnionErr":
		inj.Union = true
		cfg = mod.Config.ForkForUnion()
		parsers = newUnionParsers(inj.X(), inj.Y())
		opts = call.Args[1:]

	case "Enum", "EnumErr":
		inj.Enum = true
		cfg = mod.Config.ForkForEnum()
		parsers = enumParsers{inj.X(), inj.Y()}
		opts = call.Args[2:]

		// convgen.Enum and convgen.EnumErr takes the default enum member for
		// the output as a parameter.
		unknown, err := p.ParseEnumMember(call.Args[1], inj.Y())
		if err != nil {
			errs = errors.Join(errs, err)
		} else {
			inj.EnumUnknown = unknown
		}

	default:
		panic(codefmt.Errorf(p, callee, "unexpected convgen function: %o", callee))
	}

	// Parse config
	cfg.DiscoverBySamplePkgX = inj.X().Pkg()
	cfg.DiscoverBySamplePkgY = inj.Y().Pkg()
	errs = errors.Join(errs, p.ParseConfig(&cfg, opts, parsers))
	inj.Config = cfg

	// Register into the module
	if oldFn, ok := mod.Put(inj); !ok {
		if oldInj, ok := oldFn.(Injector); ok {
			errs = errors.Join(errs, codefmt.Errorf(p, call, `duplicate %t to %t converter
	previous declaration at %b`,
				inj.X(), inj.Y(),
				oldInj))
		} else {
			errs = errors.Join(errs, codefmt.Errorf(p, call, `duplicate %t to %t converter
	previous import of %o at %b`,
				inj.X(), inj.Y(),
				oldFn, oldFn))
		}
	}

	if errs != nil {
		return Injector{}, errs
	}
	return inj, nil
}
