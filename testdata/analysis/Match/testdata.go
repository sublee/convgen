//go:build convgen

package testdata

import (
	"github.com/sublee/convgen"
)

type (
	matchX struct{ A int }
	matchY struct{ B int }
)

// simple match
var Match1 = convgen.Struct[matchX, matchY](nil,
	convgen.Match(matchX{}.A, matchY{}.B), // ok
)

// parentheses are ok
var Match2 = convgen.Struct[matchX, matchY](nil,
	convgen.Match((matchX{}.A), (matchY{}.B)), // ok
)

// pointer receiver is ok
var Match3 = convgen.Struct[matchX, matchY](nil,
	convgen.Match((*matchX)(nil).A, (*matchY)(nil).B), // ok
)

// cross type is not allowed
var Match4 = convgen.Struct[matchX, matchY](nil,
	convgen.Match(matchY{}.B, matchY{}.B), // want `field must belong to matchX{}; got matchY{}`
)

// Missing is not allowed in Match
var Match5 = convgen.Struct[matchX, matchY](nil,
	convgen.Match(matchX{}.A, nil), // want `cannot use nil for convgen.Match`
	convgen.Match(nil, matchY{}.B), // want `cannot use nil for convgen.Match`
)

// Missing is allowed in MatchSkip
var Match6 = convgen.Struct[matchX, matchY](nil,
	convgen.MatchSkip(matchX{}.A, nil), // ok
	convgen.MatchSkip(nil, matchY{}.B), // ok
)

// parentheses are ok
var Match7 = convgen.Struct[matchX, matchY](nil,
	convgen.MatchSkip((matchX{}.A), (nil)), // ok
	convgen.MatchSkip((nil), (matchY{}.B)), // ok
)

// function match
var Match8 = convgen.Struct[matchX, matchY](nil,
	convgen.MatchFunc(matchX{}.A, matchY{}.B, func(a int) int { return a }), // ok
)

// function match with error
var Match9 = convgen.StructErr[matchX, matchY](nil,
	convgen.MatchFuncErr(matchX{}.A, matchY{}.B, func(a int) (int, error) { return a, nil }), // ok
)

// function match
var Match10 = convgen.Struct[matchX, matchY](nil,
	convgen.MatchFunc(matchX{}.A, matchY{}.B, (func(int) int)(nil)), // want `cannot use nil as function`
)
