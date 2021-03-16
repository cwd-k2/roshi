package roshi

import (
	"os"
	"path/filepath"
)

type ErrRootNotFound struct {
	msg string
}

func (e ErrRootNotFound) Error() string {
	return e.msg
}

// 親を辿って `.roshi/origin` ファイルがあるディレクトリを探す
// 返ってくるエラーは ErrRootNotFound のみ (bool でよいかもしれない)
func FindRoot(path string) (string, error) {
	filename := filepath.Join(path, ORIGIN_SPEC)

	// ファイルが存在しない場合は
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// 上のディレクトリに行く
		if parent := filepath.Dir(path); parent == path {
			return "", ErrRootNotFound{"Couldn't find roshi's root directory."}
		} else {
			return FindRoot(parent)
		}
	}

	return path, nil
}
