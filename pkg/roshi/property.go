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

// root から .roshi/origin-modtime.json を読み出す
func ReadOriginModTime(root string) (MTRecord, error) {
	filename := filepath.Join(root, ORIGIN_MODTIME)
	modified := map[string]string{}

	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	if err := json.NewDecoder(fp).Decode(&modified); err != nil {
		return nil, err
	}

	return modified, nil
}

// .roshi/origin-modtime.json を上書きする
func WriteOriginModTime(root string, modified MTRecord) error {
	filename := filepath.Join(root, ORIGIN_MODTIME)

	fp, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fp.Close()

	if err := json.NewEncoder(fp).Encode(modified); err != nil {
		return err
	}

	return nil
}

// root から .roshi/derive-modtime.json を読み出す
func ReadDeriveModTime(root string) (MTRecord, error) {
	filename := filepath.Join(root, DERIVE_MODTIME)
	modified := map[string]string{}

	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	if err := json.NewDecoder(fp).Decode(&modified); err != nil {
		return nil, err
	}

	return modified, nil
}

// .roshi/derive-modtime.json を上書きする
func WriteDeriveModTime(root string, modified MTRecord) error {
	filename := filepath.Join(root, DERIVE_MODTIME)

	fp, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fp.Close()

	if err := json.NewEncoder(fp).Encode(modified); err != nil {
		return err
	}

	return nil
}

// .roshi-ignore を読む
func ReadIgnores(root string) ([]*regexp.Regexp, error) {
	filename := filepath.Join(root, ROSHI_IGNORE)
	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	ignores := make([]*regexp.Regexp, 0)

	sc := bufio.NewScanner(fp)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		// 空行と '#' から始まる行は飛ばす
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		ignores = append(ignores, regexp.MustCompile(line))
	}

	// デフォルトで ".roshi" は ignore しておく
	ignores = append(ignores, regexp.MustCompile(".roshi"))

	return ignores, nil
}
