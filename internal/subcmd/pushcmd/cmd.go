package pushcmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	. "github.com/cwd-k2/roshi/internal/pkg/confirm"
	. "github.com/cwd-k2/roshi/internal/pkg/fileutil"
	"github.com/cwd-k2/roshi/pkg/roka"
	"github.com/cwd-k2/roshi/pkg/roshi"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var CMD = &cobra.Command{
	Use:  "push",
	Args: cobra.ExactArgs(0),
	Run:  run,
}
var (
	logger = log.New(os.Stderr, "[push] ", log.LstdFlags)

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

func skip(matching *regexp.Regexp, dfile, ofile, dpath string) bool {
	// glob にはひっかかるけど matching にはかからないのは飛ばす
	if !matching.MatchString(dfile) {
		return true
	}

	// ignore のパターンに当てはまっていたら飛ばす
	for _, ignore := range ignores {
		if ignore.MatchString(dfile) || ignore.MatchString(ofile) {
			return true
		}
	}

	// ディレクトリは困りますね (glob してきたものだから PathError はなさそう)
	if info, _ := os.Stat(dpath); info.IsDir() {
		logger.Printf("%s is a directory.\n", dpath)
		return true
	}

	modified, err := record.FileModified(dpath)
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
		globpattern := roka.CreateGlobPattern(filtration.DerivePattern)
		matches, _ := filepath.Glob(filepath.Join(root, globpattern))

		matching := roka.CreateMatchingRegexp(filtration.DerivePattern)
		template := roka.CreateTemplateString(filtration.OriginPattern, filtration.DeriNumberings)

		for _, dpath := range matches {
			// 管理下のファイル名 (相対パス)
			dfile, _ := filepath.Rel(root, dpath)
			// 元ディレクトリの対応するファイルの名前
			ofile := matching.ReplaceAllString(dfile, template)
			// 元ディレクトリのファイルのフルパス
			opath := filepath.Join(origin, ofile)

			if skip(matching, dfile, ofile, dpath) {
				continue
			}

			// ofile が更新されるかどうか
			update := true

			if FileExists(opath) {
				modified, err := record.FileModified(opath)
				if err != nil {
					logger.Printf("%+v", errors.WithStack(err))
					continue
				}

				if modified {
					msg := fmt.Sprintf("overwrite %s with %s?", ofile, dfile)
					update = Confirm(msg)
				}
			} else {
				msg := fmt.Sprintf("create new file %s?", ofile)
				update = Confirm(msg)
			}

			if update {
				if err := CopyAll(dpath, opath); err != nil {
					logger.Printf("%+v", errors.WithStack(err))
					continue
				}
				fmt.Fprintf(os.Stdout, "[push] %s => %s\n", dfile, ofile)

				record.Update(dpath)
				record.Update(opath)
			}
		}
	}
}
