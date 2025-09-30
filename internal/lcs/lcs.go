// Package lcs provides functions for finding the longest common prefix and
// suffix of a slice of strings.
package lcs

import (
	"slices"
)

// CommonPrefix returns the longest common prefix of the strings in ss.
func CommonPrefix(ss []string) string {
	// This implementation is based on os.path.commonprefix in Python.
	// https://github.com/python/cpython/blob/ed24702bd0f9925908ce48584c31dfad732208b2/Lib/genericpath.py#L105
	if len(ss) == 0 {
		return ""
	}

	// Find the lexicographically smallest and largest strings in ss.
	ss = slices.Clone(ss)
	slices.Sort(ss)

	min := slices.Min(ss)
	max := slices.Max(ss)

	// The longest common prefix of min and max is the longest common prefix of
	// ss because ss is lexicographically sorted.
	for i := range []byte(min) {
		if min[i] != max[i] {
			return min[:i]
		}
	}

	// min itself is the longest common prefix.
	return min
}

// CommonSuffix returns the longest common suffix of the strings in ss.
func CommonSuffix(ss []string) string {
	ss = slices.Clone(ss)
	for i := range ss {
		s := []byte(ss[i])
		slices.Reverse(s)
		ss[i] = string(s)
	}

	reversedSuffix := []byte(CommonPrefix(ss))
	slices.Reverse(reversedSuffix)
	suffix := string(reversedSuffix)
	return suffix
}
