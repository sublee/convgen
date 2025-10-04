package lcs

import (
	"cmp"
	"slices"
	"strings"
)

// CommonWordPrefix returns the longest common prefix of the strings in ss based
// on word boundaries detected by SplitWords.
func CommonWordPrefix(ss []string) string {
	var sss [][]string
	for _, s := range ss {
		words := SplitWords(s)
		sss = append(sss, words)
	}
	return strings.Join(lcsWord(sss), "")
}

// CommonWordSuffix returns the longest common suffix of the strings in ss based
// on word boundaries detected by SplitWords.
func CommonWordSuffix(ss []string) string {
	var sss [][]string
	for _, s := range ss {
		words := SplitWords(s)
		slices.Reverse(words)
		sss = append(sss, words)
	}
	suffix := lcsWord(sss)
	slices.Reverse(suffix)
	return strings.Join(suffix, "")
}

// lcsWord implements the longest common subsequence algorithm for a slice of
// string slices.
func lcsWord(words [][]string) []string {
	if len(words) == 0 {
		return nil
	}

	cmpFn := func(a, b []string) int {
		for i := 0; i < min(len(a), len(b)); i++ {
			if cmp := cmp.Compare(a[i], b[i]); cmp != 0 {
				return cmp
			}
		}
		return 0
	}

	slices.SortFunc(words, cmpFn)
	min := slices.MinFunc(words, cmpFn)
	max := slices.MaxFunc(words, cmpFn)

	for i := range min {
		if min[i] != max[i] {
			return min[:i]
		}
	}
	return min
}

// SplitWords splits a string into words based on character transitions. It
// detects word boundaries at:
//   - Uppercase letter after lowercase letter: "getID" -> "get" + "ID"
//   - Around underscores: "send_nowait" -> "send" + "_" + "nowait"
//   - Around digits: "file2name" -> "file" + "2" + "name"
func SplitWords(s string) []string {
	var words []string
	i := 0
	for i < len(s) {
		splitted := false

		j := i + 1
		for ; j < len(s); j++ {
			var next byte
			if j != len(s)-1 {
				next = s[j+1]
			}

			if isWordBoundary(s[j-1], s[j], next) {
				words = append(words, s[i:j])
				i = j
				splitted = true
				break
			}
		}

		if !splitted {
			words = append(words, s[i:])
			break
		}
	}
	return words
}

// isWordBoundary detects word boundaries based on character transitions.
func isWordBoundary(prev, curr, next byte) bool {
	// Uppercase after lowercase (camelCase transition)
	if prev >= 'a' && prev <= 'z' && curr >= 'A' && curr <= 'Z' {
		return true
	}
	// Uppercase before lowercase (camelCase transition)
	if curr >= 'A' && curr <= 'Z' && next >= 'a' && next <= 'z' {
		return true
	}

	// Underscore after non-underscore
	if prev != '_' && curr == '_' {
		return true
	}
	// Non-underscore after underscore
	if prev == '_' && curr != '_' {
		return true
	}

	// Digit after letter
	if (prev >= 'a' && prev <= 'z' || prev >= 'A' && prev <= 'Z') && (curr >= '0' && curr <= '9') {
		return true
	}
	// Letter after digit
	if (prev >= '0' && prev <= '9') && (curr >= 'a' && curr <= 'z' || curr >= 'A' && curr <= 'Z') {
		return true
	}

	return false
}
