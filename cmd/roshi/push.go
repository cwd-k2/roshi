package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	. "github.com/cwd-k2/roshi/internal/pkg/confirm"
	. "github.com/cwd-k2/roshi/internal/pkg/fileutil"
	"github.com/cwd-k2/roshi/pkg/roka"
	"github.com/cwd-k2/roshi/pkg/roshi"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	pushcmd = &cobra.Command{
		Use:  "push",
		Args: cobra.ExactArgs(0),
		RunE: RoshiPush,
	}
	pushlog = log.New(os.Stderr, "[push] ", log.LstdFlags|log.Ltime)
)

func init() {
	cmd.AddCommand(pushcmd)
}

func RoshiPush(c *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		pushlog.Fatalf("%+v", errors.WithStack(err))
	}

	root, err := roshi.FindRoot(cwd)
	if err != nil {
		pushlog.Fatalf("%+v", errors.WithStack(err))
	}

	origin, err := roshi.ReadOriginSpec(root)
	if err != nil {
		pushlog.Fatalf("%+v", errors.WithStack(err))
	}

	patterns, err := roshi.ReadRoshiJson(root)
	if err != nil {
		pushlog.Fatalf("%+v", errors.WithStack(err))
	}

	ignores, err := roshi.ReadIgnores(root)
	if err != nil {
		pushlog.Fatalf("%+v", errors.WithStack(err))
	}

	record, err := roshi.ReadRecord(root)
	if err != nil {
		pushlog.Fatalf("%+v", errors.WithStack(err))
	}

	filtrations, err := roka.CreateFiltrations(patterns)
	if err != nil {
		pushlog.Fatalf("%+v", errors.WithStack(err))
	}

outer:
	for _, filtration := range filtrations {
		globpattern := roka.CreateGlobPattern(filtration.DerivePattern)
		matches, _ := filepath.Glob(filepath.Join(root, globpattern))

		matching := roka.CreateMatchingRegexp(filtration.DerivePattern)
		template := roka.CreateTemplateString(filtration.OriginPattern, filtration.Numberings)

		for _, dpath := range matches {
			// 管理下のファイル名 (相対パス)
			dfile, _ := filepath.Rel(root, dpath)
			// 元ディレクトリの対応するファイルの名前
			ofile := matching.ReplaceAllString(dfile, template)

			// glob にはひっかかるけど matching にはかからないのは飛ばす
			if !matching.MatchString(dfile) {
				continue
			}

			// ignore のパターンに当てはまっていたら飛ばす
			for _, ignore := range ignores {
				if ignore.MatchString(dfile) || ignore.MatchString(ofile) {
					continue outer
				}
			}

			// glob してるから存在しないということはない
			if info, err := os.Stat(dpath); err != nil {
				pushlog.Printf("%+v", errors.WithStack(err))
				continue
			} else if info.IsDir() { // ディレクトリは困りますね
				pushlog.Printf("%s is a directory.\n", dpath)
				continue
			}

			// ファイルに更新がなければ飛ばす
			if m, err := record.FileModified(dpath); err != nil {
				pushlog.Printf("%+v", errors.WithStack(err))
				continue
			} else if !m {
				continue
			}

			// ofile が更新されるかどうか
			update := false

			// 元ディレクトリのファイルのフルパス
			opath := filepath.Join(origin, ofile)
			if FileExists(opath) {
				modified, err := record.FileModified(opath)
				if err != nil {
					pushlog.Printf("%+v", errors.WithStack(err))
					continue
				}

				if modified {
					msg := fmt.Sprintf("overwrite %s with %s?", ofile, dfile)
					update = Confirm(msg)
				} else {
					update = true
				}
			} else {
				msg := fmt.Sprintf("create new file %s?", ofile)
				update = Confirm(msg)
			}

			// 全部更新
			record.Update(dpath)

			if update {
				if err := CopyAll(dpath, opath); err != nil {
					pushlog.Printf("%+v", errors.WithStack(err))
					continue
				}
				fmt.Fprintf(os.Stdout, "[push] %s => %s\n", dfile, ofile)

				record.Update(opath)
			}
		}
	}

	return nil
}
