// `.roshi.json` 関連のバリデーションをするパッケージ
package roka

import (
	"regexp"
)

var (
	numbering = regexp.MustCompile(`(#[0-9]+)`)

	numseq = regexp.MustCompile(`(#[0-9]+#[0-9]+)`)
	thereg = regexp.MustCompile(`([\[\]\{\}\<\>\*\+\?\|\^\$])`)
	parent = regexp.MustCompile(`([/|^]\.\.[/|$])`)
)
