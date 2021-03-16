package pullcmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	. "github.com/cwd-k2/roshi/internal/pkg/confirm"
	. "github.com/cwd-k2/roshi/internal/pkg/fileutil"
	"github.com/cwd-k2/roshi/pkg/roka"
	"github.com/cwd-k2/roshi/pkg/roshi"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var CMD = &cobra.Command{
	Use:  "pull",
	Args: cobra.ExactArgs(0),
	Run:  run,
}

var (
	logger = log.New(os.Stderr, "[pull] ", log.LstdFlags)

	root        string
	origin      string
	patterns    map[string]string
	ignores     []*regexp.Regexp
	record      *roshi.MTRecord
	filtrations []*roka.Filtration
)

func prerun() error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.WithStack(err)
	}

	root, err = roshi.FindRoot(cwd)
	if err != nil {
		return errors.WithStack(err)
	}

	origin, err = roshi.ReadOriginSpec(root)
	if err != nil {
		return errors.WithStack(err)
	}

	patterns, err = roshi.ReadRoshiJson(root)
	if err != nil {
		return errors.WithStack(err)
	}

	ignores, err = roshi.ReadIgnores(root)
	if err != nil {
		return errors.WithStack(err)
	}

	record, err = roshi.ReadRecord(root)
	if err != nil {
		return errors.WithStack(err)
	}

	filtrations, err = roka.CreateFiltrations(patterns)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func skip(matching *regexp.Regexp, ofile, dfile, opath string) bool {
	// glob にはひっかかるけど matching にはかからないのは飛ばす
	if !matching.MatchString(ofile) {
		return true
	}

	// ignore のパターンに当てはまっていたら飛ばす
	for _, ignore := range ignores {
		if ignore.MatchString(ofile) || ignore.MatchString(dfile) {
			return true
		}
	}

	// ディレクトリは困りますね (glob してきたものだから PathError はなさそう)
	if info, _ := os.Stat(opath); info.IsDir() {
		logger.Printf("%s is a directory.\n", opath)
		return true
	}

	modified, err := record.FileModified(opath)
	if err != nil {
		logger.Printf("%+v", errors.WithStack(err))
		return true
	}

	// ファイルに更新がなければ飛ばす
	return !modified
}

func run(c *cobra.Command, args []string) {
	if err := prerun(); err != nil {
		logger.Fatalf("%+v", errors.WithStack(err))
	}

	for _, filtration := range filtrations {
		globpattern := roka.CreateGlobPattern(filtration.OriginPattern)
		matches, _ := filepath.Glob(filepath.Join(origin, globpattern))

		matching := roka.CreateMatchingRegexp(filtration.OriginPattern)
		template := roka.CreateTemplateString(filtration.DerivePattern, filtration.OrigNumberings)

		for _, opath := range matches {
			// 元ディレクトリのファイル名 (相対パス)
			ofile, _ := filepath.Rel(origin, opath)
			// 管理下の対応するファイルの名前
			dfile := matching.ReplaceAllString(ofile, template)
			// 管理下のファイルのフルパス
			dpath := filepath.Join(root, dfile)

			if skip(matching, ofile, dfile, opath) {
				continue
			}

			// dfile が更新されるかどうか
			update := true

			if FileExists(dpath) {
				modified, err := record.FileModified(dpath)
				if err != nil {
					logger.Printf("%+v", errors.WithStack(err))
					continue
				}

				if modified {
					msg := fmt.Sprintf("overwrite %s with %s?", dfile, ofile)
					update = Confirm(msg)
				}
			}

			if update {
				if err := CopyAll(opath, dpath); err != nil {
					logger.Printf("%+v", errors.WithStack(err))
					continue
				}
				fmt.Fprintf(os.Stdout, "[pull] %s => %s\n", ofile, dfile)

				record.Update(opath)
				record.Update(dpath)
			} else {
				timestamp := time.Now().Format("20060102030405")
				CopyAll(opath, dpath+"~"+timestamp)
				fmt.Fprintf(os.Stdout, "[pull] see %s for change.\n", dfile+"~"+timestamp)
			}

		}
	}
}
