package match

import (
	"go/token"
	"strings"

	"github.com/emirpasic/gods/maps/linkedhashmap"
	"golang.org/x/tools/go/packages"

	"github.com/sublee/convgen/internal/codefmt"
	"github.com/sublee/convgen/internal/convgen/parse"
)

type (
	renameFunc     = func(string, string) string
	findCommonFunc = func([]string) string

	object interface {
		CrumbName() string
		DebugName() string
		Pos() token.Pos
	}
)

// Matcher matches Xs to Ys by key or by forced name. It also validates whether
// the matches satisfy the expectations (missing or specific matches).
type Matcher[T object] struct {
	x, y object
	pkg  *packages.Package
	pos  token.Pos

	xs, ys *indexer[T]

	unknownPosY token.Pos
	forced      *bidiMultiMap[token.Pos, token.Pos] // posXs <-> posYs
	forcedAt    map[[2]token.Pos]token.Pos          // [posX, posY] -> where convgen.Match is called
	skippedAt   *linkedhashmap.Map                  // [posX, posY] -> where convgen.MatchSkip is called in order

	renamersX, renamersY           []renameFunc
	commonFindersX, commonFindersY []findCommonFunc
}

// NewMatcher creates a new matcher based on the given injector and config. The
// necessary options of the given config is cached in the matcher for  later
// use.
func NewMatcher[T object](inj parse.Injector, cfg parse.Config, x, y object) *Matcher[T] {
	m := &Matcher[T]{
		x:   x,
		y:   y,
		pkg: inj.Pkg(),
		pos: inj.Pos(),

		xs: newIndexer[T](),
		ys: newIndexer[T](),

		forced:    newBidiMultiMap[token.Pos, token.Pos](),
		forcedAt:  make(map[[2]token.Pos]token.Pos, len(cfg.Match)),
		skippedAt: linkedhashmap.New(),

		renamersX:      cfg.RenamersX,
		renamersY:      cfg.RenamersY,
		commonFindersX: cfg.CommonFindersX,
		commonFindersY: cfg.CommonFindersY,
	}
	for i, pair := range cfg.Match {
		pathX, pathY := pair[0], pair[1]
		m.Force(pathX.Pos, pathY.Pos, cfg.MatchAt[i])
	}
	for i, pair := range cfg.MatchSkip {
		pathX, pathY := pair[0], pair[1]
		m.Skip(pathX.Pos, pathY.Pos, cfg.MatchSkipAt[i])
	}
	return m
}

// Pkg returns the package of X and Y to satisfy [codefmt.Pkger].
func (m *Matcher[T]) Pkg() *packages.Package { return m.pkg }

// Pos returns the position of X and Y to satisfy [codefmt.Posser].
func (m *Matcher[T]) Pos() token.Pos { return m.pos }

// AddX adds an X with the given name, key, and value.
func (m *Matcher[T]) AddX(obj T, key string) { m.xs.add(obj, key) }

// AddY adds a Y with the given name, key, and value.
func (m *Matcher[T]) AddY(obj T, key string) { m.ys.add(obj, key) }

// DeleteX deletes an X with the given name.
func (m *Matcher[T]) DeleteX(pos token.Pos) { m.xs.delete(pos) }

// DeleteY deletes a Y with the given name.
func (m *Matcher[T]) DeleteY(pos token.Pos) { m.ys.delete(pos) }

// UnknownY marks the given name as the default value of Y.
// NOTE: We don't need UnknownX because X is always known.
func (m *Matcher[T]) SetUnknownY(pos token.Pos) { m.unknownPosY = pos }

func (m *Matcher[T]) Force(posX, posY, at token.Pos) {
	m.forced.Add(posX, posY)
	m.forcedAt[[2]token.Pos{posX, posY}] = at
}

func (m *Matcher[T]) Skip(posX, posY, at token.Pos) {
	m.skippedAt.Put([2]token.Pos{posX, posY}, at)
}

// Match represents a matched pair of X and Y.
type Match[T any] struct {
	X, Y T
}

// Match matches Xs and Ys. If there are any validation errors, it returns them.
func (m *Matcher[T]) Match() ([]Match[T], error) {
	matches, vis := m.matchVisualize()

	if !vis.IsValid() {
		var b strings.Builder
		for _, line := range strings.SplitAfter(vis.String(), "\n") {
			b.WriteString("\t")
			b.WriteString(line)
		}
		return nil, codefmt.Errorf(m, m, "invalid match between %s and %s\n%s", m.x.DebugName(), m.y.DebugName(), b.String())
	}

	return matches, nil
}

func (m *Matcher[T]) Visualize() string {
	_, vis := m.matchVisualize()
	return vis.String()
}

func (m *Matcher[T]) matchVisualize() ([]Match[T], *visualizer) {
	xs := m.xs.build(m.renamersX, m.commonFindersX)
	ys := m.ys.build(m.renamersY, m.commonFindersY)
	ln := newLinks()
	vis := newVisualizer()

	// Apply matching and validation rules (order matters)
	m.ruleMatch(xs, ys, ln, vis)
	m.ruleForced(xs, ys, ln, vis)
	m.ruleMissing(xs, ys, ln, vis)
	m.ruleSkip(xs, ys, ln, vis)
	m.ruleAmbiguous(xs, ys, ln, vis)

	matches := make([]Match[T], 0)
	for _, x := range xs.All {
		for _, y := range ln.FromX(x) {
			matches = append(matches, Match[T]{X: x.object.(T), Y: y.object.(T)})
		}
	}
	return matches, vis
}

// ruleMatch links X and Y by key when neither side has a forced match.
func (m *Matcher[T]) ruleMatch(xs, ys index, ln *links, vis *visualizer) {
	for _, x := range xs.All {
		if len(m.forced.Get(x.Pos())) != 0 {
			continue
		}
		for _, y := range ys.ByKey[x.key] {
			if len(m.forced.GetKeys(y.Pos())) != 0 {
				continue
			}

			// No forced match, so link by key
			ln.Link(x, y)
			vis.Match(x, y, "")
		}
	}
}

// ruleForced links X and Y according to explicit force directives.
func (m *Matcher[T]) ruleForced(xs, ys index, ln *links, vis *visualizer) {
	for posX, posY := range m.forced.All() {
		x := xs.ByPos[posX]
		y := ys.ByPos[posY]

		ln.Link(x, y)

		pos := m.forcedAt[[2]token.Pos{posX, posY}]
		reason := codefmt.Sprintf(m, "forced at %b", pos)
		vis.Match(x, y, reason)
	}
}

// ruleMissing classifies unmatched pairs:
// - convgen.MatchSkip(convgen.Missing) -> ok: skip missing
// - convgen.Enum(mod, y) -> ok: unknown value may be missing
// - otherwise -> FAIL: missing
func (m *Matcher[T]) ruleMissing(xs, ys index, ln *links, vis *visualizer) {
	for _, x := range xs.All {
		if len(ln.FromX(x)) == 0 {
			if pos, ok := m.skippedAt.Get([2]token.Pos{x.Pos(), token.NoPos}); ok {
				reason := codefmt.Sprintf(m, "skipped missing at %b", pos)
				vis.Skip(x, missing, reason)
			} else {
				vis.MatchFail(x, missing, "missing")
			}
		}
	}
	for _, y := range ys.All {
		if len(ln.FromY(y)) == 0 {
			if pos, ok := m.skippedAt.Get([2]token.Pos{token.NoPos, y.Pos()}); ok {
				reason := codefmt.Sprintf(m, "skipped missing at %b", pos)
				vis.Skip(missing, y, reason)
			} else if y.Pos() == m.unknownPosY {
				vis.Match(missing, y, "unknown value may be missing")
			} else {
				vis.MatchFail(missing, y, "missing")
			}
		}
	}
}

// ruleSkip marks matched pairs as skipped when a skip directive exists. If a
// skip is declared but the pair is not actually matched, report a failure.
func (m *Matcher[T]) ruleSkip(xs, ys index, ln *links, vis *visualizer) {
	for _, pair := range m.skippedAt.Keys() {
		posX, posY := pair.([2]token.Pos)[0], pair.([2]token.Pos)[1]
		pos, _ := m.skippedAt.Get(pair)

		x := xs.ByPos[posX]
		y := ys.ByPos[posY]

		if !ln.Linked(x, y) {
			reason := codefmt.Sprintf(m, "ineffective skip at %b", pos)
			vis.SkipFail(x, y, reason)
			continue
		}

		if !posX.IsValid() || !posY.IsValid() {
			// Already handled in "Skipped missing"
			continue
		}

		// Unlink skipped matches
		ln.Unlink(x, y)

		reason := codefmt.Sprintf(m, "skipped match at %b", pos)
		vis.Skip(x, y, reason)
	}
}

// ruleAmbiguous marks ruleAmbiguous matches as failures. Single Y linked from multiple
// Xs is ruleAmbiguous.
func (m *Matcher[T]) ruleAmbiguous(xs, ys index, ln *links, vis *visualizer) {
	for _, y := range ys.All {
		for i, x := range ln.FromY(y) {
			if i != 0 {
				// If single Y is linked from multiple Xs, it's ambiguous
				// because there is information loss. However, single X can link
				// to multiple Ys.
				vis.MatchFail(x, y, "ambiguous")
			}
		}
	}
}
