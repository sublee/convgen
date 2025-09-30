package convgenerrors_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sublee/convgen/pkg/convgenerrors"
)

func TestNil(t *testing.T) {
	err := convgenerrors.Wrap("", nil)
	assert.Nil(t, err)
}

func TestPrefix0(t *testing.T) {
	err := convgenerrors.Wrap("", errors.New("original error"))
	assert.Equal(t, "converting: original error", err.Error())
}

func TestPrefix1(t *testing.T) {
	err := convgenerrors.Wrap("Foo", errors.New("original error"))
	assert.Equal(t, "converting Foo: original error", err.Error())
}

func TestPrefix2(t *testing.T) {
	err := convgenerrors.Wrap("Foo.Bar", errors.New("original error"))
	assert.Equal(t, "converting Foo.Bar: original error", err.Error())
}

func TestPrefixFold(t *testing.T) {
	err := convgenerrors.Wrap("Foo.Bar", errors.New("original error"))
	err = convgenerrors.Wrap("Baz.Qux", err)
	assert.Equal(t, "converting Baz.Qux.Bar: original error", err.Error())
}

func TestPrefixFoldNoDot(t *testing.T) {
	err := convgenerrors.Wrap("Foo.Bar", errors.New("original error"))
	err = convgenerrors.Wrap("Baz", err)
	err = convgenerrors.Wrap("Qux", err)
	assert.Equal(t, "converting Qux.Bar: original error", err.Error())
}

func TestPrefixFoldLeadingDot(t *testing.T) {
	err := convgenerrors.Wrap("Foo.Bar", errors.New("original error"))
	err = convgenerrors.Wrap(".Baz", err)
	err = convgenerrors.Wrap("Qux", err)
	assert.Equal(t, "converting Qux.Baz.Bar: original error", err.Error())
}

func TestPrefixChainSplit(t *testing.T) {
	err := convgenerrors.Wrap("Foo.Bar", errors.New("original error"))
	err = fmt.Errorf("additional context: %w", err)
	err = convgenerrors.Wrap("Baz.Qux", err)
	assert.Equal(t, "converting Baz.Qux: additional context: converting Foo.Bar: original error", err.Error())
}

func TestErrorf(t *testing.T) {
	err := convgenerrors.Wrap("Foo.Bar", fmt.Errorf("Hello: %w", errors.New("world")))
	assert.Equal(t, "converting Foo.Bar: Hello: world", err.Error())
}

func TestErrorIs(t *testing.T) {
	orig := errors.New("original error")
	err := convgenerrors.Wrap("Foo.Bar", orig)
	assert.ErrorIs(t, err, orig)
}

type MyError struct{}

func (MyError) Error() string { return "my error" }

func TestErrorAs(t *testing.T) {
	err := convgenerrors.Wrap("Foo.Bar", MyError{})
	assert.ErrorAs(t, err, &MyError{})
}
