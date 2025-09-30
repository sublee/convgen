package codefmt

import (
	"go/ast"
	"go/token"
	"go/types"
	"io"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

// Writer is a writer for generated code.
type Writer struct {
	w       io.Writer
	pkg     *packages.Package
	fmt     Formatter
	imports map[string]Import
	ns      NS
}

// NewWriter creates a new [Writer]. It does not initialize the
// namespace. To specify a namespace, use [SetNamespace].
func NewWriter(w io.Writer, pkg *packages.Package) *Writer {
	return &Writer{
		w:       w,
		pkg:     pkg,
		fmt:     New(pkg),
		imports: make(map[string]Import),
		ns:      nil,
	}
}

// Write implements io.Writer.
func (w *Writer) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

// Printf writes a formatted string to the underlying writer using
// [CodeFormatter.Fprintf].
func (w *Writer) Printf(format string, args ...any) (int, error) {
	w.importArgs(args...)
	return w.fmt.Fprintf(w.w, format, args...)
}

// Sprintf creates a formatted string using [CodeFormatter.Sprintf].
func (w *Writer) Sprintf(format string, args ...any) string {
	w.importArgs(args...)
	return w.fmt.Sprintf(format, args...)
}

// Name returns a unique name in the namespace of the writer.
func (w *Writer) Name(name string) string {
	return w.ns.Name(name)
}

// Reserve marks a name as used in the namespace of the writer.
func (w *Writer) Reserve(name string) bool {
	return w.ns.Reserve(name)
}

// WithBuf copies the writer and sets a new write buffer.
func (w *Writer) WithBuf(buf io.Writer) *Writer {
	return &Writer{
		w:       buf,
		pkg:     w.pkg,
		fmt:     w.fmt,
		imports: w.imports,
		ns:      w.ns,
	}
}

// WithNS copies the writer and sets a new namespace.
func (w *Writer) WithNS(ns NS) *Writer {
	return &Writer{
		w:       w.w,
		pkg:     w.pkg,
		fmt:     w.fmt,
		imports: w.imports,
		ns:      ns,
	}
}

type Import struct {
	// The package to import.
	*types.Package

	// HasAlias indicates that the import has an alias.
	HasAlias bool
}

// Imports returns the collected imports. Imports are collected by [Ref] and
// [Type].
func (w *Writer) Imports() map[string]Import {
	return w.imports
}

// importAST records packages used in the given AST node to import later.
func (w *Writer) importAST(node ast.Node) {
	astutil.Apply(node, func(c *astutil.Cursor) bool {
		if id, ok := c.Node().(*ast.Ident); ok {
			w.importType(w.pkg.TypesInfo.TypeOf(id))
			w.importObj(w.pkg.TypesInfo.ObjectOf(id))
		}
		return true
	}, nil)
}

// importType records a package where the type is defined to import later.
func (w *Writer) importType(typ types.Type) {
	switch typ := typ.(type) {
	case *types.Pointer:
		w.importType(typ.Elem())
	case *types.Named:
		w.importObj(typ.Obj())
	}
}

// importObj records a package where the object is defined to import later.
func (w *Writer) importObj(obj types.Object) {
	if obj == nil {
		return
	}

	pkg := obj.Pkg()
	if pkg == nil {
		// Skip built-in objects
		return
	}

	if w.pkg.PkgPath == pkg.Path() {
		// Do not import the same package
		return
	}

	for name := range DisambiguateName(pkg.Name()) {
		prev, ok := w.imports[name]
		if ok && prev.Package == pkg {
			// Already imported with the same name.
			return
		}
		if !ok && w.pkg.Types.Scope().Lookup(name) == nil {
			// There's no conflict. Import the package with its original name.
			w.imports[name] = Import{Package: pkg, HasAlias: name != pkg.Name()}
			pkg.SetName(name)
			return
		}
	}
}

// Import adds an import for the package with the given path and alias. It
// returns the name of the imported package. The name might be different if it
// has tried to resolve name conflicts.
//
//	// fmtName can be used to refer to the "fmt" package without any name conflict.
//	fmtName := w.Import("fmt", "fmt")
//	w.Printf("%s.Println(\"Hello, World!\")", fmtName)
//
// When calling it, the package to import is recorded. Call [Imports] to
// retrieve them.
func (w *Writer) Import(path, name string) string {
	var pkgName string
	for _, imp := range w.pkg.Types.Imports() {
		if imp.Path() == path {
			pkgName = imp.Name()
			break
		}
	}

	if name == "" {
		name = pkgName
	}
	pkg := types.NewPackage(path, name)

	for name := range DisambiguateName(name) {
		prev, ok := w.imports[name]
		if ok && prev.Path() == path {
			// Already imported with the same name.
			return name
		}
		if !ok && w.pkg.Types.Scope().Lookup(name) == nil {
			w.imports[name] = Import{Package: pkg, HasAlias: name != pkgName}
			pkg.SetName(name)
			return name
		}
	}

	panic("unreachable")
}

func (w *Writer) importArgs(args ...any) {
	for _, arg := range args {
		switch arg := arg.(type) {
		case ast.Expr:
			w.importAST(arg)
		case types.Object:
			w.importObj(arg)
		case types.Type:
			w.importType(arg)

		case Exprer:
			w.importAST(arg.Expr())
		case Objecter:
			w.importObj(arg.Object())
		case Typer:
			w.importType(arg.Type())
		}
	}
}

// RewriteImports modifies the given AST node to rewrite imported package names
// to ensure there is no name conflict.
func RewriteImports[T ast.Node](w *Writer, node T) T {
	return astutil.Apply(node, func(c *astutil.Cursor) bool {
		switch node := c.Node().(type) {

		// Unqualified identifiers, such as "Println" without the "fmt." prefix
		case *ast.Ident:
			obj := w.pkg.TypesInfo.ObjectOf(node)
			if obj == nil {
				return false
			}

			pkg := obj.Pkg()
			if pkg == nil || pkg.Path() == w.pkg.PkgPath || obj.Parent() != pkg.Scope() {
				return true
			}

			newPkgName := w.Import(pkg.Path(), pkg.Name())
			c.Replace(&ast.SelectorExpr{
				X: &ast.Ident{
					NamePos: node.NamePos,
					Name:    newPkgName,
				},
				Sel: &ast.Ident{
					NamePos: node.NamePos + token.Pos(len(newPkgName)+1),
					Name:    node.Name,
					Obj:     node.Obj,
				},
			})
			return false

		// Qualified identifiers, such as "fmt.Println"
		case *ast.SelectorExpr:
			pkgIdent, ok := node.X.(*ast.Ident)
			if !ok {
				return true
			}

			pkgName, ok := w.pkg.TypesInfo.ObjectOf(pkgIdent).(*types.PkgName)
			if !ok {
				// The qualifier is not a package name.
				return true
			}

			pkg := pkgName.Imported()
			newPkgName := w.Import(pkg.Path(), pkg.Name())
			c.Replace(&ast.SelectorExpr{
				X: &ast.Ident{
					NamePos: pkgIdent.NamePos,
					Name:    newPkgName,
					Obj:     pkgIdent.Obj,
				},
				Sel: &ast.Ident{
					NamePos: pkgIdent.NamePos + token.Pos(len(pkgIdent.Name)+1),
					Name:    node.Sel.Name,
					Obj:     node.Sel.Obj,
				},
			})
			return false
		}

		// Continue traversing the AST.
		return true
	}, nil).(T)
}
