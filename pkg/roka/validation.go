package roka

import (
	"sort"
)

// ナンバリングの連続や正規表現のような危ないパターンを探す
func HasUnsafePattern(str string) bool {
	return numseq.MatchString(str) || thereg.MatchString(str) || parent.MatchString(str)
}

// スライスの中に重複した要素があるかどうか
func HasDuplicates(arr []string) bool {
	// default: false (zero value)
	dup := make(map[string]bool, 0)

	for _, pat := range arr {
		if dup[pat] {
			return true
		} else {
			dup[pat] = true
		}
	}

	return false
}

// 順序を無視してちょうど同じスライスかどうか
func SameAsSets(a, b []string) bool {
	l := len(a)

	if l != len(b) {
		return false
	}

	c := make([]string, l)
	d := make([]string, l)

	copy(c, a)
	copy(d, b)

	sort.Strings(c)
	sort.Strings(d)

	for i := 0; i < l; i++ {
		if c[i] != d[i] {
			return false
		}
	}

	return true
}

// 同じようなパターンと見られるものがダブっているかどうか
func HasDoublingPatterns(arr []string) bool {
	dbl := make(map[string]bool, 0)

	for _, pat := range arr {
		reduced := numbering.ReplaceAllString(pat, "")
		if dbl[reduced] {
			return true
		} else {
			dbl[reduced] = true
		}
	}

	return false
}
