//go:build convgen

package testdata

import "github.com/sublee/convgen"

type (
	X struct{ A Y }
	Y struct{ B int }
)

func f(...any) X { return X{} }

var Case1 = convgen.Struct[X, X](nil,
	convgen.Match(X{}.A, X{}.A.B),             // ok
	convgen.Match(X{}.A, X{}),                 // want `cannot use X itself as a field`
	convgen.Match(X{}.A, X{Y{1}}.A.B),         // want `field should belong to X\{\} with empty braces; got X\{Y\{1\}\}`
	convgen.Match(X{}.A, Y{}.B),               // want `field should belong to X\{\}; got Y\{\}`
	convgen.Match(X{}.A, struct{ A Y }{}.A.B), // want `field cannot belong to anonymous struct`

	convgen.Match(X{}.A, (*X)(nil).A.B),             // ok
	convgen.Match(X{}.A, f().A.B),                   // want `field should belong to \(\*X\)\(nil\); got f\(\)`
	convgen.Match(X{}.A, f(nil).A.B),                // want `field should belong to \(\*X\)\(nil\); got f\(nil\)`
	convgen.Match(X{}.A, f(1).A.B),                  // want `field should belong to \(\*X\)\(nil\); got f\(1\)`
	convgen.Match(X{}.A, (*Y)(nil).B),               // want `field should belong to \(\*X\)\(nil\); got \(\*Y\)\(nil\)`
	convgen.Match(X{}.A, (*struct{ A Y })(nil).A.B), // want `field cannot belong to anonymous struct`

	convgen.Match(X{}.A, (&X{}).A.B),    // ok
	convgen.Match(X{}.A, &(&X{}).A.B),   // ok
	convgen.Match(X{}.A, (*(&X{})).A.B), // ok
)

var Case2 = convgen.Struct[X, struct{ X }](nil,
	convgen.Match(X{}.A, struct{ X }{}.X.A.B), // want `field cannot belong to anonymous struct`
)

type Z struct {
	C *Y
	D []int
}

var Case3 = convgen.Struct[X, Z](nil,
	convgen.Match(X{}.A, Z{}.C.B),    // ok
	convgen.Match(X{}.A, (*Z{}.C).B), // ok

	convgen.Match(X{}.A, Z{}.D[0]),  // want `cannot use [] operation for field; got Z{}.D[0]`
	convgen.Match(X{}.A, Z{}.C.B+1), // want `cannot use \+ operation for field; got Z{}.C.B \+ 1`
	convgen.Match(X{}.A, +Z{}.C.B),  // want `cannot use \+ operation for field; got \+Z{}.C.B`
)
