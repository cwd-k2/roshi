package main

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/cwd-k2/roshi/internal/pkg/confirm"
	. "github.com/cwd-k2/roshi/internal/pkg/fileutil"
	"github.com/cwd-k2/roshi/pkg/roka"
	"github.com/cwd-k2/roshi/pkg/roshi"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var pushcmd = &cobra.Command{
	Use:  "push",
	Args: cobra.ExactArgs(0),
	RunE: RoshiPush,
}

func init() {
	cmd.AddCommand(pushcmd)
}

func RoshiPush(c *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.WithStack(err)
	}

	root, err := roshi.FindRoot(cwd)
	if err != nil {
		return errors.WithStack(err)
	}

	origin, err := roshi.ReadOriginSpec(root)
	if err != nil {
		return errors.WithStack(err)
	}

	patterns, err := roshi.ReadRoshiJson(root)
	if err != nil {
		return errors.WithStack(err)
	}

	ignores, err := roshi.ReadIgnores(root)
	if err != nil {
		return errors.WithStack(err)
	}

	omod, err := roshi.ReadOriginModTime(root)
	if err != nil {
		return errors.WithStack(err)
	}

	dmod, err := roshi.ReadDeriveModTime(root)
	if err != nil {
		return errors.WithStack(err)
	}

	filtrations, err := roka.CreateFiltrations(patterns)
	if err != nil {
		return errors.WithStack(err)
	}

outer:
	for _, filtration := range filtrations {
		globpattern := roka.CreateGlobPattern(filtration.DerivePattern)
		matches, err := filepath.Glob(filepath.Join(root, globpattern))
		if err != nil {
			return errors.WithStack(err)
		}

		matching := roka.CreateMatchingRegexp(filtration.DerivePattern)
		template := roka.CreateTemplateString(filtration.OriginPattern, filtration.Numberings)

		for _, dpath := range matches {
			// glob してるから存在しないということはない
			info, err := os.Stat(dpath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%+v", errors.WithStack(err))
				continue
			}
			// ディレクトリは困りますね
			if info.IsDir() {
				fmt.Fprintf(os.Stderr, "%s is a directory.\n", dpath)
				continue
			}
			// 管理下のファイル名 (相対パス)
			dfile, err := filepath.Rel(root, dpath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%+v", errors.WithStack(err))
				continue
			}
			// glob にはひっかかるけど matching にはかからないのは飛ばす
			if !matching.MatchString(dfile) {
				continue
			}
			// ignore のパターンに当てはまっていたら飛ばす
			for _, ignore := range ignores {
				if ignore.MatchString(dfile) {
					continue outer
				}
			}

			// 管理下のファイルの更新時刻
			dtime := info.ModTime().Format("20060102030405")

			// ファイルに更新がなければ飛ばす
			if !dmod.FileModified(dfile, dtime) {
				continue
			}

			// 元ディレクトリの対応するファイルの名前
			ofile := matching.ReplaceAllString(dfile, template)

			// ofile が更新されるかどうか
			update := false

			// 元ディレクトリのファイルのフルパス
			opath := filepath.Join(origin, ofile)
			if FileExists(opath) {
				oi, _ := os.Stat(opath)
				otime := oi.ModTime().Format("20060102030405")

				if omod.FileModified(ofile, otime) {
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
			dmod[dfile] = dtime

			if update {
				if err := CopyAll(dpath, opath); err != nil {
					fmt.Fprintf(os.Stderr, "%+v", errors.WithStack(err))
					continue
				}
				fmt.Fprintf(os.Stdout, "[push] %s => %s\n", dfile, ofile)

				oi, _ := os.Stat(opath)
				otime := oi.ModTime().Format("20060102030405")

				omod[ofile] = otime
			}
		}
	}

	if err := roshi.WriteOriginModTime(root, omod); err != nil {
		return errors.WithStack(err)
	}
	if err := roshi.WriteDeriveModTime(root, dmod); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
