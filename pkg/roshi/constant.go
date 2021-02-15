package roshi

import "path/filepath"

const (
	sep          string = string(filepath.Separator)
	ROSHI_DIR    string = ".roshi"
	ROSHI_JSON   string = ".roshi.json"
	ROSHI_IGNORE string = ".roshi-ignore"
	ROSHI_OBJECT string = ROSHI_DIR + sep + "object"
	ORIGIN_SPEC  string = ROSHI_DIR + sep + "origin"
)
