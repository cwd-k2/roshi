package roka

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// マッチする正規表現と置換先のテンプレートとなる文字列を作るための道具
type Filtration struct {
	OriginPattern  string
	DerivePattern  string
	OrigNumberings []string
	DeriNumberings []string
}

func CreateFiltrations(obj map[string]string) ([]*Filtration, error) {
	var (
		filtrations = make([]*Filtration, 0)
		originstrs  = make([]string, 0)
		derivestrs  = make([]string, 0)
	)

	// 個々のパターンを見る
	for origin, derive := range obj {
		// 危ないパターン
		if HasUnsafePattern(origin) {
			errstr := fmt.Sprintf("%s has unsafe pattern.", origin)
			return nil, errors.New(errstr)
		}
		if HasUnsafePattern(derive) {
			errstr := fmt.Sprintf("%s has unsafe pattern.", derive)
			return nil, errors.New(errstr)
		}

		// ナンバリングに重複があるパターン
		arr1 := numbering.FindAllString(origin, -1)
		if arr1 != nil && HasDuplicates(arr1) {
			errstr := fmt.Sprintf("%s has duplicated numberings.", origin)
			return nil, errors.New(errstr)
		}
		arr2 := numbering.FindAllString(derive, -1)
		if arr2 != nil && HasDuplicates(arr2) {
			errstr := fmt.Sprintf("%s has duplicated numberings.", derive)
			return nil, errors.New(errstr)
		}

		// origin と derive で十分に対応がとれない組
		if !SameAsSets(arr1, arr2) {
			errstr := fmt.Sprintf("Connot make corresponding patterning with %s and %s.", origin, derive)
			return nil, errors.New(errstr)
		}

		originstrs = append(originstrs, origin)
		derivestrs = append(derivestrs, derive)

		// 文字列の長さを長い順にしないと #1 と #10 とかが上手く置換できないので必要
		// origin, derive でナンバリングが出てくる順番も重要であるため, 両方の持っておく必要がある
		var orignumberings []string
		orignumberings = append(orignumberings, arr1...)
		sort.SliceStable(orignumberings, func(i, j int) bool {
			return len(orignumberings[i]) > len(orignumberings[j])
		})

		var derinumberings []string
		derinumberings = append(derinumberings, arr2...)
		sort.SliceStable(derinumberings, func(i, j int) bool {
			return len(derinumberings[i]) > len(derinumberings[j])
		})

		filtration := &Filtration{
			OriginPattern:  origin,
			DerivePattern:  derive,
			OrigNumberings: orignumberings,
			DeriNumberings: derinumberings,
		}
		filtrations = append(filtrations, filtration)
	}

	// origin, derive 全体について同一とみなせるようなパターンの対があるとアウト
	// ここで判定するのは非効率な気がするけど後で考える
	if HasDoublingPatterns(originstrs) {
		return nil, errors.New("Doubled patterns in origin patterns.")
	}
	if HasDoublingPatterns(derivestrs) {
		return nil, errors.New("Doubled patterns in derive patterns.")
	}

	return filtrations, nil
}

// マッチするパスを探すための正規表現を作る
func CreateMatchingRegexp(base string) *regexp.Regexp {
	safedot := strings.ReplaceAll(base, `.`, `\.`)
	replreg := numbering.ReplaceAllString(safedot, `([^/\.]+?)`)
	return regexp.MustCompile(fmt.Sprintf("^%s$", replreg))
}

// 正規表現ではなく Glob 用のパターンを作る
func CreateGlobPattern(base string) string {
	return numbering.ReplaceAllString(base, "*")
}

// マッチしたものを置換する先の文字列を作る
func CreateTemplateString(base string, numberings []string) string {
	for i, pattern := range numberings {
		base = strings.ReplaceAll(base, pattern, fmt.Sprintf("${%d}", i+1))
	}
	return base
}
