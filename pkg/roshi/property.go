package roshi

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// root から origin を読み出す
func ReadOriginSpec(root string) (string, error) {
	originBS, err := ioutil.ReadFile(filepath.Join(root, ORIGIN_SPEC))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(originBS)), nil
}

// root から .roshi.json を map[string]string として読み出す
func ReadRoshiJson(root string) (map[string]string, error) {
	filename := filepath.Join(root, ROSHI_JSON)
	patterns := map[string]string{}

	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	if err := json.NewDecoder(fp).Decode(&patterns); err != nil {
		return nil, err
	}

	return patterns, nil
}

// root から .roshi/object の内容を読み書きする構造体を作成する
func ReadRecord(root string) (*MTRecord, error) {
	dirname := filepath.Join(root, ROSHI_OBJECT)

	if err := os.MkdirAll(dirname, os.ModePerm); err != nil {
		return nil, err
	}

	mtrecord := &MTRecord{dirname}

	return mtrecord, nil
}

// .roshi-ignore を読む
func ReadIgnores(root string) ([]*regexp.Regexp, error) {
	filename := filepath.Join(root, ROSHI_IGNORE)
	// デフォルトで `.roshi` は ignore しておく
	ignores := []*regexp.Regexp{regexp.MustCompile(`\.roshi`)}

	// .roshi.ignore が無い場合は無視
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return ignores, nil
	}

	fp, err := os.Open(filename)
	if err != nil {
		return ignores, err
	}
	defer fp.Close()

	sc := bufio.NewScanner(fp)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		// 空行と '#' から始まる行は飛ばす
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		ignores = append(ignores, regexp.MustCompile(line))
	}

	return ignores, nil
}
