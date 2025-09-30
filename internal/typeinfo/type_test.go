package typeinfo_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sublee/convgen/internal/typeinfo"
)

func parse(code string) (*ast.File, *types.Info, *types.Package, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", code, parser.AllErrors)
	if err != nil {
		return nil, nil, nil, err
	}

	info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}
	pkg, err := (&types.Config{}).Check("pkg", fset, []*ast.File{file}, info)
	if err != nil {
		return nil, nil, nil, err
	}

	return file, info, pkg, nil
}

func parseType(typeExpr string) (types.Type, error) {
	_, _, pkg, err := parse(fmt.Sprintf("package p; var x %s", typeExpr))
	if err != nil {
		return nil, err
	}
	x := pkg.Scope().Lookup("x")
	return x.Type(), nil
}

func TestTypeIdentical(t *testing.T) {
	ty1, err := parseType("int")
	require.NoError(t, err)

	ty2, err := parseType("int")
	require.NoError(t, err)

	ti1 := typeinfo.TypeOf(ty1)
	ti2 := typeinfo.TypeOf(ty2)
	assert.True(t, ti1.Identical(ti2))
	assert.True(t, ti2.Identical(ti1))
}

func TestTypeNotIdentical(t *testing.T) {
	ty1, err := parseType("int")
	require.NoError(t, err)

	ty2, err := parseType("string")
	require.NoError(t, err)

	ti1 := typeinfo.TypeOf(ty1)
	ti2 := typeinfo.TypeOf(ty2)
	assert.False(t, ti1.Identical(ti2))
	assert.False(t, ti2.Identical(ti1))
}

func TestTypeOfBasic(t *testing.T) {
	ty, err := parseType("int")
	require.NoError(t, err)

	ti := typeinfo.TypeOf(ty)
	assert.True(t, ti.IsBasic())
}

func TestTypeOfArray(t *testing.T) {
	ty, err := parseType("[3]int")
	require.NoError(t, err)

	ti := typeinfo.TypeOf(ty)
	assert.True(t, ti.IsArray())
	assert.True(t, ti.Elem.IsBasic())
	assert.Equal(t, int64(3), ti.Len)
}

func TestTypeOfSlice(t *testing.T) {
	ty, err := parseType("[]int")
	require.NoError(t, err)

	ti := typeinfo.TypeOf(ty)
	assert.True(t, ti.IsSlice())
	assert.True(t, ti.Elem.IsBasic())
}

func TestTypeOfMap(t *testing.T) {
	ty, err := parseType("map[int]int")
	require.NoError(t, err)

	ti := typeinfo.TypeOf(ty)
	assert.True(t, ti.IsMap())
	assert.True(t, ti.Elem.IsBasic())
	assert.True(t, ti.Key.IsBasic())
}

func TestTypeOfStruct(t *testing.T) {
	ty, err := parseType("struct{ x int}")
	require.NoError(t, err)

	ti := typeinfo.TypeOf(ty)
	assert.True(t, ti.IsStruct())

	x, ok := ti.StructField("x")
	require.True(t, ok)

	tiX := typeinfo.TypeOf(x.Type())
	assert.True(t, tiX.IsBasic())
}

func TestTypeOfInterface(t *testing.T) {
	ty, err := parseType("interface{}")
	require.NoError(t, err)

	ti := typeinfo.TypeOf(ty)
	assert.True(t, ti.IsInterface())
}

func TestTypeOfPointer(t *testing.T) {
	ty, err := parseType("*int")
	require.NoError(t, err)

	ti := typeinfo.TypeOf(ty)
	assert.True(t, ti.IsPointer())
	assert.True(t, ti.Elem.IsBasic())
	assert.True(t, ti.PointerDepth() == 1)
	assert.True(t, ti.Deref().IsBasic())
}

func TestTypeOfPointer2(t *testing.T) {
	ty, err := parseType("**int")
	require.NoError(t, err)

	ti := typeinfo.TypeOf(ty)
	assert.True(t, ti.IsPointer())
	assert.True(t, ti.Elem.IsPointer())
	assert.True(t, ti.Elem.Elem.IsBasic())
	assert.True(t, ti.PointerDepth() == 2)
	assert.True(t, ti.Deref().IsBasic())
}

func TestTypeOfNamed(t *testing.T) {
	_, _, pkg, err := parse(`
package p
type myInt int
var x myInt
`)
	require.NoError(t, err)

	ty := pkg.Scope().Lookup("x").Type()

	ti := typeinfo.TypeOf(ty)
	assert.True(t, ti.IsNamed())
	assert.True(t, ti.IsBasic())
}

func TestTypeOfError(t *testing.T) {
	ty, err := parseType("error")
	require.NoError(t, err)

	ti := typeinfo.TypeOf(ty)
	assert.True(t, ti.IsError())
}

func TestTypeOfNil(t *testing.T) {
	file, info, _, err := parse(`
package p
func f(x any) int { return 0 }
var x = f(nil)
`)
	require.NoError(t, err)

	arg := file.Decls[1].(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Values[0].(*ast.CallExpr).Args[0]
	ty := info.TypeOf(arg)

	ti := typeinfo.TypeOf(ty)
	assert.True(t, ti.IsNil())
}

func TestTypeOfGeneric(t *testing.T) {
	file, info, _, err := parse(`
package p
type A[T, U any] struct{ x T; y U }
type B[U any] A[int, U]
type C A[int, int]
`)
	require.NoError(t, err)

	nthTypeExpr := func(n int) ast.Expr {
		return file.Decls[n].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type
	}

	tyA := info.TypeOf(nthTypeExpr(0))
	tiA := typeinfo.TypeOf(tyA)
	assert.True(t, tiA.IsGeneric())

	tyB := info.TypeOf(nthTypeExpr(1))
	tiB := typeinfo.TypeOf(tyB)
	assert.True(t, tiB.IsGeneric())

	tyC := info.TypeOf(nthTypeExpr(2))
	tiC := typeinfo.TypeOf(tyC)
	assert.False(t, tiC.IsGeneric())
}

func TestTypeRef(t *testing.T) {
	ty, err := parseType("int")
	require.NoError(t, err)
	ti := typeinfo.TypeOf(ty)

	ref := ti.Ref()
	assert.True(t, ref.IsPointer())
	assert.True(t, ref.Elem.Identical(ti))
}
