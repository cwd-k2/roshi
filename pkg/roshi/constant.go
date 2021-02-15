package roshi

import "path/filepath"

const (
	sep            string = string(filepath.Separator)
	ROSHI_DIR      string = ".roshi"
	ROSHI_JSON     string = ".roshi.json"
	ROSHI_IGNORE   string = ".roshi-ignore"
	ORIGIN_SPEC    string = ROSHI_DIR + sep + "origin"
	ORIGIN_MODTIME string = ROSHI_DIR + sep + "origin-modtime.json"
	DERIVE_MODTIME string = ROSHI_DIR + sep + "derive-modtime.json"
)
