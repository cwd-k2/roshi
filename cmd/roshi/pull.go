package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	. "github.com/cwd-k2/roshi/internal/pkg/confirm"
	. "github.com/cwd-k2/roshi/internal/pkg/fileutil"
	"github.com/cwd-k2/roshi/pkg/roka"
	"github.com/cwd-k2/roshi/pkg/roshi"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	pullcmd = &cobra.Command{
		Use:  "pull",
		Args: cobra.ExactArgs(0),
		RunE: RoshiPull,
	}
	pulllog = log.New(os.Stderr, "[pull] ", log.LstdFlags)
)

func init() {
	cmd.AddCommand(pullcmd)
}

func RoshiPull(c *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		pulllog.Fatalf("%+v", errors.WithStack(err))
	}

	root, err := roshi.FindRoot(cwd)
	if err != nil {
		pulllog.Fatalf("%+v", errors.WithStack(err))
	}

	origin, err := roshi.ReadOriginSpec(root)
	if err != nil {
		pulllog.Fatalf("%+v", errors.WithStack(err))
	}

	patterns, err := roshi.ReadRoshiJson(root)
	if err != nil {
		pulllog.Fatalf("%+v", errors.WithStack(err))
	}

	ignores, err := roshi.ReadIgnores(root)
	if err != nil {
		pulllog.Fatalf("%+v", errors.WithStack(err))
	}

	record, err := roshi.ReadRecord(root)
	if err != nil {
		pulllog.Fatalf("%+v", errors.WithStack(err))
	}

	filtrations, err := roka.CreateFiltrations(patterns)
	if err != nil {
		pulllog.Fatalf("%+v", errors.WithStack(err))
	}

outer:
	for _, filtration := range filtrations {
		globpattern := roka.CreateGlobPattern(filtration.OriginPattern)
		matches, _ := filepath.Glob(filepath.Join(origin, globpattern))

		matching := roka.CreateMatchingRegexp(filtration.OriginPattern)
		template := roka.CreateTemplateString(filtration.DerivePattern, filtration.Numberings)

		for _, opath := range matches {
			// 元ディレクトリのファイル名 (相対パス)
			ofile, _ := filepath.Rel(origin, opath)
			// 管理下の対応するファイルの名前
			dfile := matching.ReplaceAllString(ofile, template)

			// glob にはひっかかるけど matching にはかからないのは飛ばす
			if !matching.MatchString(ofile) {
				continue
			}

			// ignore のパターンに当てはまっていたら飛ばす
			for _, ignore := range ignores {
				if ignore.MatchString(ofile) || ignore.MatchString(dfile) {
					continue outer
				}
			}

			// ディレクトリは困りますね (glob してきたものだから PathError はなさそう)
			if info, _ := os.Stat(opath); info.IsDir() {
				pulllog.Printf("%s is a directory.\n", opath)
				continue
			}

			// ファイルに更新がなければ飛ばす
			if m, err := record.FileModified(opath); err != nil {
				pulllog.Printf("%+v", errors.WithStack(err))
				continue
			} else if !m {
				continue
			}

			// dfile が更新されるかどうか
			update := false

			// 管理下のファイルのフルパス
			dpath := filepath.Join(root, dfile)
			if FileExists(dpath) {
				modified, err := record.FileModified(dpath)
				if err != nil {
					pulllog.Printf("%+v", errors.WithStack(err))
					continue
				}

				if modified {
					msg := fmt.Sprintf("overwrite %s with %s?", dfile, ofile)
					update = Confirm(msg)

					if !update {
						timestamp := time.Now().Format("20060102030405")
						CopyAll(opath, dpath+"~"+timestamp)
						fmt.Fprintf(os.Stdout, "[pull] see %s for change.\n", dfile+"~"+timestamp)
					}

				} else {
					update = true
				}
			} else {
				update = true
			}

			// 全部更新
			record.Update(opath)

			if update {
				if err := CopyAll(opath, dpath); err != nil {
					pulllog.Printf("%+v", errors.WithStack(err))
					continue
				}
				fmt.Fprintf(os.Stdout, "[pull] %s => %s\n", ofile, dfile)

				record.Update(dpath)
			}

		}
	}

	return nil
}
