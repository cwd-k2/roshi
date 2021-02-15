package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cwd-k2/roshi/pkg/roshi"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	initcmd = &cobra.Command{
		Use:  "init /path/to/origin",
		Args: cobra.ExactArgs(1),
		RunE: RoshiInit,
	}
	initlog = log.New(os.Stderr, "[init] ", log.LstdFlags)
)

func init() {
	cmd.AddCommand(initcmd)
}

func RoshiInit(c *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		initlog.Fatalf("%+v", errors.WithStack(err))
	}

	srcdir, err := filepath.Abs(args[0])
	if err != nil {
		initlog.Fatalf("%+v", errors.WithStack(err))
	}

	// TODO: srcdir が cwd 以下にないことを保証

	// .roshi のルートを探してみる (存在しないことが望まれる)
	if dir, err := roshi.FindRoot(cwd); err == nil { // 既に roshi の管理下にある場合
		initlog.Printf("A directory %s is already initialized\n", dir)
		return nil
	} else if _, ok := err.(roshi.ErrRootNotFound); !ok { // ErrRootNotFound 以外のエラーの場合
		initlog.Fatalf("%+v", errors.WithStack(err))
	}

	// .roshi/ を作成する
	if err := os.MkdirAll(filepath.Join(cwd, roshi.ROSHI_DIR), os.ModePerm); err != nil && !os.IsExist(err) {
		initlog.Fatalf("%+v", errors.WithStack(err))
	}

	// .roshi/origin (text file)
	if err := CreateRoshiOrigin(cwd, srcdir); err != nil {
		initlog.Fatalf("%+v", errors.WithStack(err))
	}

	// .roshi.json を作成 (存在しなければ)
	if err := CreateRoshiJson(cwd); err != nil {
		initlog.Fatalf("%+v", errors.WithStack(err))
	}

	return nil
}

func CreateRoshiOrigin(p, srcdir string) error {
	fp, err := os.Create(filepath.Join(p, roshi.ORIGIN_SPEC))
	if err != nil {
		return err
	}

	if _, err := strings.NewReader(srcdir).WriteTo(fp); err != nil {
		return err
	}

	if err := fp.Close(); err != nil {
		return err
	}

	return nil
}

func CreateRoshiJson(p string) error {
	filename := filepath.Join(p, roshi.ROSHI_JSON)

	// ファイルが存在しない以外の場合 (nil なら存在してるので nil 返せる)
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		return err
	}

	// 空の .roshi.json を作成
	fp, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fp.Close()

	if err := json.NewEncoder(fp).Encode(map[string]string{}); err != nil {
		return err
	}

	return nil
}
