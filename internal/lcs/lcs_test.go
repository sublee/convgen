package lcs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sublee/convgen/internal/lcs"
)

func TestCommonPrefix(t *testing.T) {
	ss := []string{"prefix", "prefill", "present"}
	got := lcs.CommonPrefix(ss)
	assert.Equal(t, "pre", got)
}

func TestCommonPrefixItself(t *testing.T) {
	ss := []string{"hello", "hell", "hel"}
	got := lcs.CommonPrefix(ss)
	assert.Equal(t, "hel", got)
}

func TestCommonPrefixEmpty(t *testing.T) {
	ss := []string{}
	got := lcs.CommonPrefix(ss)
	assert.Equal(t, "", got)
}

func TestNoCommonPrefix(t *testing.T) {
	ss := []string{"dependency", "feel", "extend"}
	got := lcs.CommonPrefix(ss)
	assert.Equal(t, "", got)
}

func TestCommonPrefixUnicode(t *testing.T) {
	ss := []string{"안경", "안돼", "안녕"}
	got := lcs.CommonPrefix(ss)
	assert.Equal(t, "안", got)
}

func TestCommonSuffix(t *testing.T) {
	ss := []string{"awful", "beautiful", "colorful"}
	got := lcs.CommonSuffix(ss)
	assert.Equal(t, "ful", got)
}

func TestCommonSuffixItself(t *testing.T) {
	ss := []string{"hello", "ello", "llo"}
	got := lcs.CommonSuffix(ss)
	assert.Equal(t, "llo", got)
}

func TestCommonSuffixEmpty(t *testing.T) {
	ss := []string{}
	got := lcs.CommonSuffix(ss)
	assert.Equal(t, "", got)
}

func TestNoCommonSuffix(t *testing.T) {
	ss := []string{"happy", "smile", "8mile"}
	got := lcs.CommonSuffix(ss)
	assert.Equal(t, "", got)
}

func TestCommonSuffixUnicode(t *testing.T) {
	ss := []string{"건물", "동물", "식물"}
	got := lcs.CommonSuffix(ss)
	assert.Equal(t, "물", got)
}
