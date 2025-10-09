package convgeninternal

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"io"
	"maps"
	"path/filepath"
	"slices"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"

	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/convgen/assign"
	"github.com/sublee/convgen/internal/convgen/parse"
)

// Convgen generates converter code for the target package. Call [Build] and
// then [Generate] to get the generated code. All potential errors are returned
// by [Build]. Once [Build] succeeds, [Generate] never fails.
type Convgen struct {
	p   *parse.Parser
	ns  codefmt.NS
	buf *bytes.Buffer
	w   *codefmt.Writer

	mods     map[token.Pos]*parse.Module
	injs     map[token.Pos]parse.Injector
	convs    map[token.Pos]assign.Conv
	subconvs map[*parse.Module][]assign.Conv
}

// New creates a new [Convgen] for the given package. If the package does not
// satisfy the requirements, an error is returned. The package must have its
// Syntax, Types and TypesInfo. And it must not have any errors.
func New(pkg *packages.Package) (*Convgen, error) {
	parser, err := parse.New(pkg)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	return &Convgen{
		p:        parser,
		ns:       codefmt.NewNS(pkg.Types.Scope()),
		buf:      &buf,
		w:        codefmt.NewWriter(&buf, pkg),
		subconvs: make(map[*parse.Module][]assign.Conv),
	}, nil
}

// Build prepares code generation by parsing code and building converters. All
// potential errors are returned by this method. It must be called before
// [Generate].
func (cg *Convgen) Build() error {
	// Parse modules and injectors from the package.
	mods, errs := cg.p.ParseModules()
	cg.mods = mods

	errs = errors.Join(errs, cg.p.Validate(mods))

	injs, err := cg.p.ParseInjectors(cg.ns, mods)
	errs = errors.Join(errs, err)

	if errs != nil {
		return errs
	}
	if len(injs) == 0 {
		// No converter definitions found
		return nil
	}

	cg.injs = make(map[token.Pos]parse.Injector)
	for _, inj := range injs {
		cg.injs[inj.Pos()] = inj
	}

	// Initialize the namespace with all function names in the modules to
	// reserve them against conflicts.
	for _, inj := range injs {
		cg.ns.Reserve(inj.Name())

		for fn := range inj.Module.Range() {
			if fn.Name() != "" && fn.Object().Pkg() == cg.p.Pkg().Types {
				// Reserve imported name of functions defined in the current
				// package.
				cg.ns.Reserve(fn.Name())
			}
		}
	}

	// Build converters from the definitions.
	cg.convs = make(map[token.Pos]assign.Conv)
	for _, inj := range injs {
		conv, subconvs, err := assign.Build(inj, cg.ns, cg.subconvs[inj.Module])
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		cg.convs[inj.Pos()] = conv
		cg.subconvs[inj.Module] = append(cg.subconvs[inj.Module], subconvs...)
	}

	return errs
}

// Generate generates converter code for the package. It must be called after
// [Build] succeeds.
func (cg *Convgen) Generate() []byte {
	cg.writeConvCode()
	cg.mergeCode()
	return cg.frameCode()
}

// writeConvCode writes function declaration code for explicit and implicit
// converters.
func (cg *Convgen) writeConvCode() {
	if len(cg.convs) != 0 {
		cg.w.Printf("// convgen: explicit converters\n\n")

		convs := slices.Collect(maps.Values(cg.convs))
		slices.SortFunc(convs, func(a, b assign.Conv) int {
			if a.Pos() < b.Pos() {
				return -1
			}
			if a.Pos() > b.Pos() {
				return 1
			}
			return 0
		})

		for _, conv := range convs {
			local := maps.Clone(cg.ns)
			w := cg.w.WithNS(local)
			conv.WriteDefineCode(w)
			cg.w.Printf("\n")
		}
	}

	subconvCount := 0
	for _, subconvs := range cg.subconvs {
		subconvCount += len(subconvs)
	}
	if subconvCount != 0 {
		cg.w.Printf("// convgen: implicit converters\n\n")

		sortedMods := slices.Collect(maps.Keys(cg.subconvs))
		slices.SortFunc(sortedMods, func(a, b *parse.Module) int {
			if a.Name < b.Name {
				return -1
			}
			if a.Name > b.Name {
				return 1
			}
			return 0
		})

		for _, mod := range sortedMods {
			for _, conv := range cg.subconvs[mod] {
				local := maps.Clone(cg.ns)
				w := cg.w.WithNS(local)
				conv.WriteDefineCode(w)
				cg.w.Printf("\n")
			}
		}
	}
}

// mergeCode copies non-convgen code from the source files that tagged with
// "//go:build convgen". It erases convgen directives to remove any references
// to the convgen package.
func (cg *Convgen) mergeCode() {
	for _, file := range cg.p.ConvgenGoFiles() {
		name := filepath.Base(cg.p.Pkg().Fset.File(file.Pos()).Name())
		first := true

		for _, decl := range file.Decls {
			if gen, ok := decl.(*ast.GenDecl); ok {
				if gen.Tok == token.IMPORT {
					// Skip import declarations in files. Required imports will
					// be collected from their usage, and then rewritten as an
					// import declaration group.
					continue
				}
			}

			if first {
				fmt.Fprintf(cg.buf, "// %s:\n\n", name)
				first = false
			}

			// Erase convgen.Module()
			decl = astutil.Apply(decl, func(c *astutil.Cursor) bool {
				if call, ok := c.Node().(*ast.CallExpr); ok {
					if cg.p.IsDirective(call, "Module") {
						// HACK: printer.Fprint does not validate the name of an
						// Ident node. It can be used to inject arbitrary code
						// including comments at the desired position.
						c.Replace(&ast.Ident{Name: "struct{}{} // convgen module erased"})
						return false
					}
				}
				return true
			}, nil).(ast.Decl)

			// Erase converter injectors
			decl = astutil.Apply(decl, func(c *astutil.Cursor) bool {
				spec, ok := c.Node().(*ast.ValueSpec)
				if !ok {
					return true
				}

				// Find non-convgen values
				var names []*ast.Ident
				var values []ast.Expr
				for i := range spec.Names {
					if i >= len(spec.Values) {
						// Enum consts may not have values
						names = append(names, spec.Names[i])
						continue
					}

					if _, ok := cg.convs[spec.Values[i].Pos()]; !ok {
						names = append(names, spec.Names[i])
						values = append(values, spec.Values[i])
					}
				}

				if len(names) == 0 {
					// Input:  var ( a = convgen.Struct(...) )
					// Output: var ()
					c.Delete()
				} else {
					// Input:  var ( a, b = convgen.Struct(...), 42 )
					// Output: var ( b = 42 )
					c.Replace(&ast.ValueSpec{
						Doc:     spec.Doc,
						Names:   names,
						Type:    spec.Type,
						Values:  values,
						Comment: spec.Comment,
					})
				}

				return false
			}, nil).(ast.Decl)

			// Skip empty declarations
			if gen, ok := decl.(*ast.GenDecl); ok {
				if len(gen.Specs) == 0 {
					continue
				}
			}

			// Prevent import name conflicts when merging multiple files into one
			decl = codefmt.RewriteImports(cg.w, decl)

			// Write rewritten declaration code
			printer.Fprint(cg.buf, cg.p.Pkg().Fset, &printer.CommentedNode{
				Node:     decl,
				Comments: file.Comments,
			})
			fmt.Fprintf(cg.buf, "\n\n")
		}
	}
}

func (cg *Convgen) frameCode() []byte {
	// Prepend header code
	versionSuffix := ""
	if Version != "" {
		versionSuffix = "@" + Version
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "//go:build !convgen\n")
	fmt.Fprintf(&buf, "// Code generated by github.com/sublee/convgen%s. DO NOT EDIT.\n", versionSuffix)
	fmt.Fprintf(&buf, "package %s\n", cg.p.Pkg().Name)

	if len(cg.w.Imports()) != 0 {
		fmt.Fprintf(&buf, "import (\n")
		for alias, imp := range cg.w.Imports() {
			// Check for remaining convgen import
			if imp.Path() == "github.com/sublee/convgen" {
				fmt.Println("convgen import remains")
			}

			if imp.HasAlias {
				fmt.Fprintf(&buf, "%s %q\n", alias, imp.Path())
			} else {
				fmt.Fprintf(&buf, "%q\n", imp.Path())
			}
		}
		fmt.Fprintf(&buf, ")\n")
	}

	_, _ = io.Copy(&buf, cg.buf)
	code := buf.Bytes()

	// Apply gofmt if succeeded
	if fmtCode, err := format.Source(code); err == nil {
		code = fmtCode
	}
	return code
}
