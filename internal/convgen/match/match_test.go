package match_test

import (
	"go/token"
	"go/types"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sublee/convgen/internal/convgen/match"
	"github.com/sublee/convgen/internal/convgen/parse"
	"github.com/sublee/convgen/internal/typeinfo"
)

// anInj is a dummy injector that does nothing. Passed to NewMatcher to satisfy
// the interface.
var anInj = parse.Injector{
	Func: typeinfo.NewFunc(
		nil,
		"",
		typeinfo.TypeOf(types.Universe.Lookup("nil").Type()),
		typeinfo.TypeOf(types.Universe.Lookup("nil").Type()),
		false,
		false,
	),
}

type Obj struct {
	n  int
	cn string
}

func (o Obj) CrumbName() string { return o.cn }
func (o Obj) DebugName() string { return o.cn }
func (o Obj) Pos() token.Pos    { return token.Pos(o.n) }

var dummy = Obj{999, "DUMMY"}

// ss: standardize space
func ss(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func Test1to1(t *testing.T) {
	m := match.NewMatcher[Obj](anInj, parse.Config{}, dummy, dummy)
	m.AddX(Obj{1, "fruit.apple"}, "A")
	m.AddY(Obj{2, "person.alice"}, "A")
	m.AddX(Obj{3, "fruit.banana"}, "B")
	m.AddY(Obj{4, "person.bob"}, "B")

	matches, err := m.Match()
	require.NoError(t, err)

	require.Len(t, matches, 2)
	assert.Equal(t, 1, matches[0].X.n)
	assert.Equal(t, 2, matches[0].Y.n)
	assert.Equal(t, 3, matches[1].X.n)
	assert.Equal(t, 4, matches[1].Y.n)
}

func TestMissing(t *testing.T) {
	m := match.NewMatcher[Obj](anInj, parse.Config{}, dummy, dummy)
	m.AddX(Obj{1, "fruit.apple"}, "Apple")
	m.AddY(Obj{2, "person.alice"}, "Alice")
	m.AddX(Obj{3, "fruit.banana"}, "Banana")
	m.AddY(Obj{4, "person.bob"}, "Bob")
	m.AddX(Obj{5, "fruit.clementine"}, "Clementine")
	m.AddY(Obj{6, "person.clementine"}, "Clementine")

	v := m.Visualize()
	assert.Contains(t, ss(v), ss(`
FAIL: Apple [apple]           -> ? // missing
FAIL: Banana [banana]         -> ? // missing
ok:   Clementine [clementine] -> Clementine [clementine]
FAIL: ?                       -> Alice [alice] // missing
FAIL: ?                       -> Bob [bob]     // missing
`), v)
}

func firstChar(s, _ string) string { return s[:1] }

func TestRename(t *testing.T) {
	var cfg parse.Config
	cfg.RenamersX = append(cfg.RenamersX, firstChar)
	cfg.RenamersY = append(cfg.RenamersY, firstChar)
	cfg.CommonFindersX = append(cfg.CommonFindersX, nil)
	cfg.CommonFindersY = append(cfg.CommonFindersY, nil)

	m := match.NewMatcher[Obj](anInj, cfg, dummy, dummy)
	m.AddX(Obj{1, "fruit.apple"}, "Apple")
	m.AddY(Obj{2, "person.alice"}, "Alice")
	m.AddX(Obj{3, "fruit.banana"}, "Banana")
	m.AddY(Obj{4, "person.bob"}, "Bob")

	matches, err := m.Match()
	require.NoError(t, err)

	require.Len(t, matches, 2)
	assert.Equal(t, 1, matches[0].X.n)
	assert.Equal(t, 2, matches[0].Y.n)
	assert.Equal(t, 3, matches[1].X.n)
	assert.Equal(t, 4, matches[1].Y.n)
}

func Test1to2(t *testing.T) {
	var cfg parse.Config
	cfg.RenamersX = append(cfg.RenamersX, firstChar)
	cfg.RenamersY = append(cfg.RenamersY, firstChar)
	cfg.CommonFindersX = append(cfg.CommonFindersX, nil)
	cfg.CommonFindersY = append(cfg.CommonFindersY, nil)

	m := match.NewMatcher[Obj](anInj, cfg, dummy, dummy)
	m.AddX(Obj{1, "fruit.apple"}, "Apple")
	m.AddY(Obj{2, "person.alice"}, "Alice")
	m.AddY(Obj{3, "person.adam"}, "Adam")

	matches, err := m.Match()
	require.NoError(t, err)

	require.Len(t, matches, 2)
	assert.Equal(t, 1, matches[0].X.n)
	assert.Equal(t, 2, matches[0].Y.n)
	assert.Equal(t, 1, matches[1].X.n)
	assert.Equal(t, 3, matches[1].Y.n)
}

func Test2to1_Ambiguous(t *testing.T) {
	var cfg parse.Config
	cfg.RenamersX = append(cfg.RenamersX, firstChar)
	cfg.RenamersY = append(cfg.RenamersY, firstChar)
	cfg.CommonFindersX = append(cfg.CommonFindersX, nil)
	cfg.CommonFindersY = append(cfg.CommonFindersY, nil)

	m := match.NewMatcher[Obj](anInj, cfg, dummy, dummy)
	m.AddX(Obj{1, "fruit.apple"}, "Apple")
	m.AddX(Obj{2, "fruit.avocado"}, "Avocado")
	m.AddX(Obj{3, "fruit.acai"}, "AcaiBerry")
	m.AddY(Obj{4, "person.alice"}, "Alice")

	v := m.Visualize()
	assert.Contains(t, ss(v), ss(`
ok:   A [apple]   -> A [alice]
FAIL: A [avocado] -> A [alice] // ambiguous
FAIL: A [acai]    -> A [alice] // ambiguous
`), v)
}

func TestForced(t *testing.T) {
	m := match.NewMatcher[Obj](anInj, parse.Config{}, dummy, dummy)
	m.AddX(Obj{1, "fruit.apple"}, "A")
	m.AddY(Obj{2, "person.alice"}, "A")
	m.AddX(Obj{3, "fruit.banana"}, "B")
	m.AddY(Obj{4, "person.bob"}, "B")

	m.Force(1, 4, token.NoPos)
	m.Force(3, 2, token.NoPos)

	v := m.Visualize()
	assert.Contains(t, ss(v), ss(`
ok: A [apple]  -> B [bob]   // forced at -:-
ok: B [banana] -> A [alice] // forced at -:-
`), v)
}

func TestSkipped(t *testing.T) {
	m := match.NewMatcher[Obj](anInj, parse.Config{}, dummy, dummy)
	m.AddX(Obj{1, "fruit.apple"}, "A")
	m.AddY(Obj{2, "person.alice"}, "A")
	m.AddX(Obj{3, "fruit.banana"}, "B")
	m.AddY(Obj{4, "person.bob"}, "B")
	m.AddX(Obj{5, "fruit.durian"}, "D")
	m.AddY(Obj{6, "person.clementine"}, "C")

	m.Skip(1, 2, token.NoPos)
	m.Skip(3, 2, token.NoPos)
	m.Skip(5, 0, token.NoPos)
	m.Skip(0, 6, token.NoPos)

	v := m.Visualize()
	assert.Contains(t, ss(v), ss(`
ok:   A [apple]  .. A [alice]      // skipped match at -:-
FAIL: B [banana] .. A [alice]      // ineffective skip at -:-
ok:   B [banana] -> B [bob]
ok:   D [durian] .. ?              // skipped missing at -:-
ok:   ?          .. C [clementine] // skipped missing at -:-
`), v)
}

func TestDefault(t *testing.T) {
	m := match.NewMatcher[Obj](anInj, parse.Config{}, dummy, dummy)
	m.AddX(Obj{1, "fruit.apple"}, "A")
	m.AddY(Obj{2, "person.alice"}, "A")
	m.AddY(Obj{3, "person.bob"}, "B")
	m.SetDefaultY(3)

	v := m.Visualize()
	assert.Contains(t, ss(v), ss(`
ok:   A [apple] -> A [alice]
ok:   ?         -> B [bob] // missing allowed as default
`), v)
}

func TestDelete(t *testing.T) {
	m := match.NewMatcher[Obj](anInj, parse.Config{}, dummy, dummy)
	m.AddX(Obj{1, "fruit.apple"}, "A")
	m.AddY(Obj{2, "person.alice"}, "A")
	m.AddX(Obj{3, "fruit.banana"}, "B")
	m.AddY(Obj{4, "person.bob"}, "B")

	m.DeleteX(1)
	m.DeleteY(2)

	assert.Equal(t, "ok: B [banana] -> B [bob]", m.Visualize())
}

func TestMatchErrorIndent(t *testing.T) {
	m := match.NewMatcher[Obj](anInj, parse.Config{}, dummy, dummy)
	m.AddX(Obj{1, "x"}, "x")
	m.AddY(Obj{2, "y"}, "y")

	matches, err := m.Match()
	assert.Len(t, matches, 0)
	assert.ErrorContains(t, err, `invalid match between DUMMY and DUMMY
	FAIL: x -> ? // missing
	FAIL: ? -> y // missing`)
}
