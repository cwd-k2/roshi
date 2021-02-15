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
func FindRoot(path string) (string, error) {
	filename := filepath.Join(path, ORIGIN_SPEC)

	// ファイルが存在しない, または `.roshi/origin` がディレクトリ (草)
	if info, err := os.Stat(filename); os.IsNotExist(err) || err == nil && info.IsDir() {
		// 上のディレクトリに行く
		if parent := filepath.Dir(path); parent == path {
			return "", ErrRootNotFound{"Couldn't find roshi's root directory."}
		} else {
			return FindRoot(parent)
		}
	} else if err != nil {
		// よくわからないエラーが出たときは終わり
		return "", err
	}

	return path, nil
}
