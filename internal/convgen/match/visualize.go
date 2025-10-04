package match

import (
	"cmp"
	"go/token"
	"io"
	"maps"
	"slices"
	"strings"
	"text/tabwriter"
)

// visualizer helps visualize the matching results and validation errors. It
// renders the results in a tabular format like below:
//
//	ok:   A -> A
//	FAIL: B -> ?   // missing
//	ok:   C -> Cat // forced at main.go:10:5
//	ok:   D .. ?   // skipped missing at main.go:11:5
type visualizer struct {
	matches map[[2]entry]validity
}

// newVisualizer creates a new visualizer.
func newVisualizer() *visualizer {
	return &visualizer{matches: make(map[[2]entry]validity)}
}

// validity represents whether a match is valid, skipped, and the reason.
type validity struct {
	ok      bool
	skipped bool
	reason  string
}

// IsValid reports whether all matches are valid.
func (vis visualizer) IsValid() bool {
	for _, v := range vis.matches {
		if !v.ok {
			return false
		}
	}
	return true
}

// put puts a match result into the visualizer.
func (vis *visualizer) put(x, y entry, val validity) {
	vis.matches[[2]entry{x, y}] = val

	if x.IsValid() && y.IsValid() {
		// If both x and y are valid, remove any previous missing entries
		// because they are no longer missing.
		delete(vis.matches, [2]entry{x, missing})
		delete(vis.matches, [2]entry{missing, y})
	}
}

// Match records a valid match.
func (vis *visualizer) Match(x, y entry, reason string) {
	vis.put(x, y, validity{ok: true, skipped: false, reason: reason})
}

// Skip records a skipped match.
func (vis *visualizer) Skip(x, y entry, reason string) {
	vis.put(x, y, validity{ok: true, skipped: true, reason: reason})
}

func (vis *visualizer) MatchFail(x, y entry, reason string) {
	vis.put(x, y, validity{ok: false, skipped: false, reason: reason})
}

// Dots records a skipped match.
func (vis *visualizer) SkipFail(x, y entry, reason string) {
	vis.put(x, y, validity{ok: false, skipped: true, reason: reason})
}

// String returns the string representation of the visualizer.
func (vis visualizer) String() string {
	var b strings.Builder
	tw := tabwriter.NewWriter(&b, 1, 1, 1, ' ', 0)

	pairs := slices.Collect(maps.Keys(vis.matches))
	slices.SortFunc(pairs, func(a, b [2]entry) int {
		if cmp := cmpValidWinsInvalid(a[0].Pos(), b[0].Pos()); cmp != 0 {
			return cmp
		}
		if cmp := cmpAboveWinsBelow(a[0].Pos(), b[0].Pos()); cmp != 0 {
			return cmp
		}
		if cmp := cmpValidWinsInvalid(a[1].Pos(), b[1].Pos()); cmp != 0 {
			return cmp
		}
		if cmp := cmpAboveWinsBelow(a[1].Pos(), b[1].Pos()); cmp != 0 {
			return cmp
		}
		return 0
	})

	first := true
	for _, pair := range pairs {
		x, y := pair[0], pair[1]
		v := vis.matches[pair]

		if first {
			first = false
		} else {
			io.WriteString(tw, "\n")
		}

		if v.ok {
			io.WriteString(tw, "ok:\t")
		} else {
			io.WriteString(tw, "FAIL:\t")
		}

		io.WriteString(tw, x.String())

		if v.skipped {
			io.WriteString(tw, "\t..\t")
		} else {
			io.WriteString(tw, "\t->\t")
		}

		io.WriteString(tw, y.String())

		if v.reason != "" {
			io.WriteString(tw, "\t// ")
			io.WriteString(tw, v.reason)
		}
	}

	tw.Flush()
	return b.String()
}

func cmpValidWinsInvalid(a, b token.Pos) int {
	if a.IsValid() && !b.IsValid() {
		return -1
	}
	if !a.IsValid() && b.IsValid() {
		return 1
	}
	return 0
}

func cmpAboveWinsBelow(a, b token.Pos) int {
	return cmp.Compare(a, b)
}
