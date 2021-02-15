package fileutil

import (
	"bufio"
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// ディレクトリごとファイルを作る
func CreateAll(filename string) (*os.File, error) {
	// ディレクトリを先に作る
	d, _ := filepath.Split(filename)
	if err := os.MkdirAll(d, os.ModePerm); err != nil {
		return nil, err
	}

	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// ファイルを新規作成して内容を全部書き込む
func CopyAll(srcpath, dstpath string) error {
	dfp, err := CreateAll(dstpath)
	if err != nil {
		return err
	}
	defer dfp.Close()

	ofp, err := os.Open(srcpath)
	if err != nil {
		return err
	}
	defer ofp.Close()

	r := bufio.NewReader(ofp)
	w := bufio.NewWriter(dfp)

	if _, err := r.WriteTo(w); err != nil {
		return err
	}

	return nil
}
