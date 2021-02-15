package roshi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
)

// ファイルのハッシュ値を読み書きする構造体
type MTRecord struct {
	directory string
}

// ファイルが更新されているかどうか
func (m *MTRecord) FileModified(f string) (bool, error) {
	p := filepath.Join(m.directory, FileNameHash(f))

	// 存在しない場合
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return true, nil
	}

	phash, err := ioutil.ReadFile(p)
	if err != nil {
		return true, err
	}
	fhash, err := FileContentHash(f)
	if err != nil {
		return true, err
	}

	// 等しくないときに更新されている
	return !bytes.Equal(phash, fhash), nil
}

// ファイルのハッシュを更新
func (m *MTRecord) Update(f string) error {
	p := filepath.Join(m.directory, FileNameHash(f))

	fp, err := os.Create(p)
	if err != nil {
		return err
	}
	defer fp.Close()

	b, err := FileContentHash(f)
	if err != nil {
		return err
	}

	_, err = fp.Write(b)

	return err
}

func FileNameHash(filename string) string {
	b := sha256.Sum256([]byte(filename))
	return hex.EncodeToString(b[:])
}

func FileContentHash(filename string) ([]byte, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(b)

	return hash[:], nil
}
