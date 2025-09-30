package typeinfo

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// Func describes a function information. It holds function signature
// information of [types.Object] that is necessary from the Convgen's
// perspective.
type Func interface {
	Object() types.Object
	Name() string          // for declared function
	FuncLit() *ast.FuncLit // for anonymous function literal

	// Signature information
	X() Type
	Y() Type
	HasErr() bool
	HasOut() bool

	// Position information
	Pos() token.Pos
	WithPos(token.Pos) Func
}

// function implements the [Func] interface.
type function struct {
	obj    types.Object
	lit    *ast.FuncLit
	x      Type
	y      Type
	hasErr bool
	hasOut bool
	pos    token.Pos
}

func (fn function) Object() types.Object { return fn.obj }
func (fn function) Name() string {
	if fn.obj == nil {
		return ""
	}
	return fn.obj.Name()
}

func (fn function) FuncLit() *ast.FuncLit { return fn.lit }

func (fn function) X() Type      { return fn.x }
func (fn function) Y() Type      { return fn.y }
func (fn function) HasErr() bool { return fn.hasErr }
func (fn function) HasOut() bool { return fn.hasOut }

func (fn function) Pos() token.Pos {
	if fn.pos == token.NoPos {
		return fn.obj.Pos()
	}
	return fn.pos
}

// NewFunc creates a new [Func] with the given attributes.
//
// This is for creating a new function programmatically. It focuses on the
// signature of the function only, not the body.
//
// To create from an existing function object, use [FuncOf] or related
// functions. To create from a function literal without a named object, use
// [InspectFuncLit] or related functions.
func NewFunc(pkg *types.Package, name string, x, y Type, hasErr, hasOut bool) Func {
	var params, results []*types.Var
	params = append(params, types.NewVar(token.NoPos, pkg, "in", x.T))
	if hasOut {
		params = append(params, types.NewVar(token.NoPos, pkg, "out", types.NewPointer(y.T)))
	} else {
		results = append(results, types.NewVar(token.NoPos, pkg, "out", y.T))
	}
	if hasErr {
		results = append(results, types.NewVar(token.NoPos, pkg, "err", types.Universe.Lookup("error").Type()))
	}
	sig := types.NewSignatureType(nil, nil, nil, types.NewTuple(params...), types.NewTuple(results...), false)

	obj := types.NewFunc(token.NoPos, pkg, name, sig)
	return function{
		obj:    obj,
		x:      x,
		y:      y,
		hasErr: hasErr,
		hasOut: hasOut,
	}
}

// WithPos returns a copy of the [Func] with the given position.
func (fn function) WithPos(pos token.Pos) Func {
	return function{fn.obj, fn.lit, fn.x, fn.y, fn.hasErr, fn.hasOut, pos}
}

// Shape is a type constraint for function shapes. It is used in [FuncOf] and
// [FuncLitOf] to specify which kind of function signature is expected.
//
//	Shape   | Basic      | HasErr              | HasOut       | HasErr & HasOut
//	--------+------------+---------------------+--------------+-------------------
//	NoXY    | func()     | func() error        | -            | -
//	OnlyX   | func(x)    | func(x) error       | -            | -
//	OnlyY   | func() y   | func() (y, error)   | func(*y),    | func(*y) error
//	BothXY  | func(x) y  | func(x) (y, error)  | func(x, *y)  | func(x, *y) error
type Shape interface {
	NoXY | OnlyX | OnlyY | BothXY
	needXY() (bool, bool)
}

type (
	// NoXY is the shape with no input and no output.
	NoXY struct{}

	// OnlyX is the shape with only input X.
	OnlyX struct{}

	// OnlyY is the shape with only output Y.
	OnlyY struct{}

	// BothXY is the shape with both input X and output Y.
	BothXY struct{}
)

func (NoXY) needXY() (bool, bool)   { return false, false }
func (OnlyX) needXY() (bool, bool)  { return true, false }
func (OnlyY) needXY() (bool, bool)  { return false, true }
func (BothXY) needXY() (bool, bool) { return true, true }

func funcOfNoXY(f function, params, results *types.Tuple) (Func, error) {
	switch {
	case params.Len() == 0 && results.Len() == 0:
		// func()
		fallthrough
	case params.Len() == 0 && results.Len() == 1 && f.hasErr:
		// func() error
		return f, nil
	}
	return nil, fmt.Errorf("expected signature: [func()], [func() error]")
}

func funcOfOnlyX(f function, params, results *types.Tuple) (Func, error) {
	switch {
	case params.Len() == 1 && results.Len() == 0:
		// func(X)
		fallthrough
	case params.Len() == 1 && results.Len() == 1 && f.hasErr:
		// func(X) error
		f.x = TypeOf(params.At(0).Type())
		return f, nil
	}
	return nil, fmt.Errorf("expected signature: [func(X)], [func(X) error]")
}

func funcOfOnlyY(f function, params, results *types.Tuple) (Func, error) {
	switch {
	case params.Len() == 0 && results.Len() == 1:
		// func() Y
		fallthrough
	case params.Len() == 0 && results.Len() == 2 && f.hasErr:
		// func() (Y, error)
		f.y = TypeOf(results.At(0).Type())
		return f, nil
	case params.Len() == 1 && (results.Len() == 0 || results.Len() == 1 && f.hasErr):
		// func(*Y) or func(*Y) error
		if ptr := TypeOf(params.At(0).Type()); ptr.IsPointer() {
			f.y = *ptr.Elem
			f.hasOut = true
			return f, nil
		}
	}
	return nil, fmt.Errorf("expected signature: [func() Y], [func() (Y, error)]")
}

func funcOfBothXY(f function, params, results *types.Tuple) (Func, error) {
	switch {
	case params.Len() == 1 && results.Len() == 1:
		// func(X) Y
		fallthrough
	case params.Len() == 1 && results.Len() == 2 && f.hasErr:
		// func(X) (Y, error)
		f.x = TypeOf(params.At(0).Type())
		f.y = TypeOf(results.At(0).Type())
		return f, nil
	case params.Len() == 2 && results.Len() == 0 && !f.hasErr:
		// func(X, *Y)
		if ptr := TypeOf(params.At(1).Type()); ptr.IsPointer() {
			f.x = TypeOf(params.At(0).Type())
			f.y = *ptr.Elem
			f.hasOut = true
			return f, nil
		}
	case params.Len() == 2 && results.Len() == 1 && f.hasErr:
		// func(X, *Y) error
		if ptr := TypeOf(params.At(1).Type()); ptr.IsPointer() {
			f.x = TypeOf(params.At(0).Type())
			f.y = *ptr.Elem
			f.hasOut = true
			return f, nil
		}
	}
	return nil, fmt.Errorf("expected signature: [func(X) Y], [func(X) (Y, error)], [func(X, *Y) bool], [func(X, *Y) error]")
}

// FuncOf inspects the given function and returns a new [Func]. It returns an
// error if the function signature does not match with the given shape.
func FuncOf[S Shape](obj types.Object) (Func, error) {
	sig, ok := obj.Type().Underlying().(*types.Signature)
	if !ok {
		return nil, fmt.Errorf("func: not signature type")
	}

	f := function{obj: obj}
	if n := sig.Results().Len(); n != 0 && isTypeError(sig.Results().At(n-1).Type()) {
		f.hasErr = true
	}

	needX, needY := S{}.needXY()
	switch {
	case !needX && !needY:
		return funcOfNoXY(f, sig.Params(), sig.Results())
	case needX && !needY:
		return funcOfOnlyX(f, sig.Params(), sig.Results())
	case !needX && needY:
		return funcOfOnlyY(f, sig.Params(), sig.Results())
	default: // needX && needY
		return funcOfBothXY(f, sig.Params(), sig.Results())
	}
}

// FuncLitOf inspects the given function literal and returns a new [Func]. It
// returns an error if the function signature does not match with the given
// shape.
func FuncLitOf[S Shape](pkg *packages.Package, lit *ast.FuncLit) (Func, error) {
	sig := pkg.TypesInfo.TypeOf(lit).(*types.Signature)
	obj := types.NewFunc(token.NoPos, pkg.Types, "", sig)

	fn, err := FuncOf[S](obj)
	if err != nil {
		return nil, err
	}

	return function{
		obj:    obj,
		lit:    lit,
		x:      fn.X(),
		y:      fn.Y(),
		hasErr: fn.HasErr(),
		hasOut: fn.HasOut(),
	}, nil
}

// isTypeError reports whether t is the built-in error type.
func isTypeError(t types.Type) bool {
	return t == types.Universe.Lookup("error").Type()
}
