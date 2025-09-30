package codefmt

import (
	"fmt"
	"go/token"
	"go/types"
	"iter"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// NS manages unique names in a namespace.
type NS map[string]struct{}

// NewNS creates a new namespace which reserves all names in the given
// scope.
func NewNS(scope *types.Scope) NS {
	ns := make(NS)
	for _, name := range scope.Names() {
		ns.Reserve(name)
	}
	return ns
}

// Reserve marks a name as used in the namespace. If the name is already used,
// it returns false.
func (ns NS) Reserve(name string) bool {
	if _, ok := ns[name]; ok {
		return false
	}
	ns[name] = struct{}{}
	return true
}

// Name returns a unique name in its namespace. Once a name is used, it is
// reserved in the namespace to avoid conflicts. If conflicts occur, a numbering
// suffix is added.
//
// Panics if the name is empty.
func (ns NS) Name(name string) string {
	name = NormalizeName(name)
	if ns == nil {
		return name
	}
	if token.Lookup(name).IsKeyword() {
		return name
	}
	for name := range DisambiguateName(name) {
		if ok := ns.Reserve(name); ok {
			return name
		}
	}
	panic("unreachable")
}

func NormalizeName(name string) string {
	if name == "" {
		panic("empty name")
	}

	chunks := strings.FieldsFunc(name, func(r rune) bool {
		return !('a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || '0' <= r && r <= '9' || r == '_')
	})

	for i := 1; i < len(chunks); i++ {
		chunks[i] = cases.Title(language.English).String(chunks[i])
	}
	return strings.Join(chunks, "")
}

// DisambiguateName offers an alternative unique names.
func DisambiguateName(name string) iter.Seq[string] {
	if name == "" {
		panic("empty name")
	}

	return func(yield func(string) bool) {
		if !yield(name) {
			return
		}

		// Postfix "_" to the name if it already ends with a number.
		// "answer42_2" is better than "answer422".
		sep := ""
		if name[len(name)-1] != '_' && name[len(name)-1] >= '0' && name[len(name)-1] <= '9' {
			sep = "_"
		}

		for i := 2; ; i++ {
			if !yield(fmt.Sprintf("%s%s%d", name, sep, i)) {
				return
			}
		}
	}
}
