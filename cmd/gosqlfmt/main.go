package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyokomi/gosqlfmt/formatter"
)

var (
	flagWrite = flag.Bool("w", false, "フォーマット結果でファイルを上書きする")
	flagDiff  = flag.Bool("d", false, "フォーマット前後のdiffを表示する")
	flagList  = flag.Bool("l", false, "フォーマットが必要なファイルパスを一覧表示する")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: gosqlfmt [flags] [path ...]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n引数なしの場合、stdinから読み込みstdoutに出力する。\n")
	}
	flag.Parse()

	exitCode := 0
	if flag.NArg() == 0 {
		if err := processStdin(&exitCode); err != nil {
			fmt.Fprintf(os.Stderr, "gosqlfmt: %v\n", err)
			os.Exit(2)
		}
	} else {
		for _, arg := range flag.Args() {
			processPath(arg, &exitCode)
		}
	}
	os.Exit(exitCode)
}

func processStdin(exitCode *int) error {
	src, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("stdin読み込みエラー: %w", err)
	}

	formatted, err := formatter.FormatSource(src)
	if err != nil {
		return fmt.Errorf("フォーマットエラー: %w", err)
	}

	if *flagDiff {
		if !bytes.Equal(src, formatted) {
			diff, err := unifiedDiff("stdin", src, formatted)
			if err != nil {
				return fmt.Errorf("diff生成エラー: %w", err)
			}
			os.Stdout.Write(diff)
			*exitCode = 1
		}
		return nil
	}

	if *flagList {
		if !bytes.Equal(src, formatted) {
			fmt.Println("<stdin>")
			*exitCode = 1
		}
		return nil
	}

	os.Stdout.Write(formatted)
	return nil
}

func processPath(path string, exitCode *int) {
	// ./... パターンのサポート
	if strings.HasSuffix(path, "/...") || strings.HasSuffix(path, string(os.PathSeparator)+"...") {
		dir := strings.TrimSuffix(path, "/...")
		dir = strings.TrimSuffix(dir, string(os.PathSeparator)+"...")
		if dir == "" || dir == "." {
			dir = "."
		}
		processDirectory(dir, exitCode)
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gosqlfmt: %v\n", err)
		*exitCode = 2
		return
	}

	if info.IsDir() {
		processDirectory(path, exitCode)
	} else {
		processFile(path, exitCode)
	}
}

func processDirectory(dir string, exitCode *int) {
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "gosqlfmt: %v\n", err)
			*exitCode = 2
			return nil
		}

		if d.IsDir() {
			name := d.Name()
			// .始まりディレクトリとvendorをスキップ
			if strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			if name == "vendor" || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		// テストファイルはスキップしない（gofmt準拠）
		processFile(path, exitCode)
		return nil
	})
}

func processFile(filename string, exitCode *int) {
	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gosqlfmt: %v\n", err)
		*exitCode = 2
		return
	}

	formatted, err := formatter.FormatSource(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gosqlfmt: %s: %v\n", filename, err)
		*exitCode = 2
		return
	}

	if bytes.Equal(src, formatted) {
		if !*flagList && !*flagDiff && !*flagWrite {
			os.Stdout.Write(formatted)
		}
		return
	}

	// 変更がある場合
	if *flagList {
		fmt.Println(filename)
		*exitCode = 1
	}

	if *flagDiff {
		diff, err := unifiedDiff(filename, src, formatted)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gosqlfmt: %s: %v\n", filename, err)
			*exitCode = 2
			return
		}
		os.Stdout.Write(diff)
		*exitCode = 1
	}

	if *flagWrite {
		info, err := os.Stat(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gosqlfmt: %s: %v\n", filename, err)
			*exitCode = 2
			return
		}
		if err := os.WriteFile(filename, formatted, info.Mode()); err != nil {
			fmt.Fprintf(os.Stderr, "gosqlfmt: %s: %v\n", filename, err)
			*exitCode = 2
			return
		}
	}

	if !*flagList && !*flagDiff && !*flagWrite {
		os.Stdout.Write(formatted)
	}
}
