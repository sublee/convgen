package codefmt_test

import (
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"

	"github.com/sublee/convgen/internal/codefmt"
)

type pkger struct{}

func (pkger) Pkg() *packages.Package {
	var pkg packages.Package
	pkg.Fset = token.NewFileSet()
	pkg.Fset.AddFile("test.go", -1, 100).AddLine(10)
	return &pkg
}

type poser struct{ pos int }

func (p poser) Pos() token.Pos { return token.Pos(p.pos) }

func TestErrorfNilNil(t *testing.T) {
	err := codefmt.Errorf(nil, nil, "simple error")
	assert.Equal(t, "simple error", err.Error())
}

func TestErrorfPos(t *testing.T) {
	err := codefmt.Errorf(pkger{}, poser{1}, "error")
	assert.Equal(t, "test.go:1:1: error", err.Error())
}

func TestErrorfW(t *testing.T) {
	assert.Panics(t, func() {
		_ = codefmt.Errorf(pkger{}, poser{1}, "error: %w", assert.AnError)
	})
}
