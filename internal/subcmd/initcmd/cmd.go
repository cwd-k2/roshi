package initcmd

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

var CMD = &cobra.Command{
	Use:  "init /path/to/origin",
	Args: cobra.ExactArgs(1),
	Run:  run,
}

var (
	logger = log.New(os.Stderr, "[init] ", log.LstdFlags)
)

func run(c *cobra.Command, args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		logger.Fatalf("%+v", errors.WithStack(err))
	}

	srcdir, err := filepath.Abs(args[0])
	if err != nil {
		logger.Fatalf("%+v", errors.WithStack(err))
	}

	// TODO: srcdir が cwd 以下にないことを保証

	// .roshi のルートを探してみる (存在しないことが望まれる)
	if dir, err := roshi.FindRoot(cwd); err == nil { // 既に roshi の管理下にある場合
		logger.Printf("A directory %s is already initialized\n", dir)
		os.Exit(0)
	}

	// .roshi/ を作成する
	if err := os.MkdirAll(filepath.Join(cwd, roshi.ROSHI_DIR), os.ModePerm); err != nil && !os.IsExist(err) {
		logger.Fatalf("%+v", errors.WithStack(err))
	}

	// .roshi/origin (text file)
	if err := CreateRoshiOrigin(cwd, srcdir); err != nil {
		logger.Fatalf("%+v", errors.WithStack(err))
	}

	// .roshi.json を作成 (存在しなければ)
	if err := CreateRoshiJson(cwd); err != nil {
		logger.Fatalf("%+v", errors.WithStack(err))
	}
}

func CreateRoshiOrigin(p, srcdir string) error {
	fp, err := os.Create(filepath.Join(p, roshi.ORIGIN_SPEC))
	if err != nil {
		return err
	}
	defer fp.Close()

	if _, err := strings.NewReader(srcdir).WriteTo(fp); err != nil {
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
