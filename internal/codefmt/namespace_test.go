package codefmt

import (
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisambiguate(t *testing.T) {
	pull, stop := iter.Pull(DisambiguateName("example"))
	defer stop()

	var name string
	var more bool

	name, more = pull()
	assert.Equal(t, "example", name)
	assert.True(t, more)

	name, more = pull()
	assert.Equal(t, "example2", name)
	assert.True(t, more)

	name, more = pull()
	assert.Equal(t, "example3", name)
	assert.True(t, more)
}

func TestDisambiguateNumSuffix(t *testing.T) {
	pull, stop := iter.Pull(DisambiguateName("answer42"))
	defer stop()

	var name string
	var more bool

	name, more = pull()
	assert.Equal(t, "answer42", name)
	assert.True(t, more)

	name, more = pull()
	assert.Equal(t, "answer42_2", name)
	assert.True(t, more)

	name, more = pull()
	assert.Equal(t, "answer42_3", name)
	assert.True(t, more)
}
