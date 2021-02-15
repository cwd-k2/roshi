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

var pullcmd = &cobra.Command{
	Use:  "pull",
	Args: cobra.ExactArgs(0),
	RunE: RoshiPull,
}

func init() {
	cmd.AddCommand(pullcmd)
}

func RoshiPull(c *cobra.Command, args []string) error {
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
		globpattern := roka.CreateGlobPattern(filtration.OriginPattern)
		matches, err := filepath.Glob(filepath.Join(origin, globpattern))
		if err != nil {
			return errors.WithStack(err)
		}

		matching := roka.CreateMatchingRegexp(filtration.OriginPattern)
		template := roka.CreateTemplateString(filtration.DerivePattern, filtration.Numberings)

		for _, opath := range matches {
			// glob してるから存在しないということはない
			info, err := os.Stat(opath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%+v", errors.WithStack(err))
				continue
			}
			// ディレクトリは困りますね
			if info.IsDir() {
				fmt.Fprintf(os.Stderr, "%s is a directory.\n", opath)
				continue
			}
			// 元ディレクトリのファイル名 (相対パス)
			ofile, err := filepath.Rel(origin, opath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%+v", errors.WithStack(err))
				continue
			}
			// glob にはひっかかるけど matching にはかからないのは飛ばす
			if !matching.MatchString(ofile) {
				continue
			}
			// ignore のパターンに当てはまっていたら飛ばす
			for _, ignore := range ignores {
				if ignore.MatchString(ofile) {
					continue outer
				}
			}

			// 元ディレクトリのファイルの更新時刻
			otime := info.ModTime().Format("20060102030405")

			// ファイルに更新がなければ飛ばす
			if !omod.FileModified(ofile, otime) {
				continue
			}

			// 管理下の対応するファイルの名前
			dfile := matching.ReplaceAllString(ofile, template)

			// dfile が更新されるかどうか
			update := false

			// 管理下のファイルのフルパス
			dpath := filepath.Join(root, dfile)
			if FileExists(dpath) {
				di, _ := os.Stat(dpath)
				dtime := di.ModTime().Format("20060102030405")

				if dmod.FileModified(dfile, dtime) {
					msg := fmt.Sprintf("overwrite %s with %s?", dfile, ofile)
					if Confirm(msg) {
						update = true
					} else {
						CopyAll(opath, dpath+"~"+otime)
						fmt.Fprintf(os.Stdout, "[pull] see %s for change.\n", dfile+"~"+otime)
					}
				} else {
					update = true
				}
			} else {
				update = true
			}

			// 全部更新
			omod[ofile] = otime

			if update {
				if err := CopyAll(opath, dpath); err != nil {
					fmt.Fprintf(os.Stderr, "%+v", errors.WithStack(err))
					continue
				}
				fmt.Fprintf(os.Stdout, "[pull] %s => %s\n", ofile, dfile)

				di, _ := os.Stat(dpath)
				dtime := di.ModTime().Format("20060102030405")

				dmod[dfile] = dtime
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
