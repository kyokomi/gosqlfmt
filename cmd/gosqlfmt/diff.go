package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

// unifiedDiff は外部diffコマンドでunified diffを生成する。
// diffコマンドが利用できない場合はフォールバック出力を返す。
func unifiedDiff(filename string, original, formatted []byte) ([]byte, error) {
	dir, err := os.MkdirTemp("", "gosqlfmt")
	if err != nil {
		return fallbackDiff(filename, original, formatted), nil
	}
	defer os.RemoveAll(dir)

	origFile := filepath.Join(dir, "original")
	fmtFile := filepath.Join(dir, "formatted")

	if err := os.WriteFile(origFile, original, 0o600); err != nil {
		return fallbackDiff(filename, original, formatted), nil
	}
	if err := os.WriteFile(fmtFile, formatted, 0o600); err != nil {
		return fallbackDiff(filename, original, formatted), nil
	}

	cmd := exec.Command("diff", "-u",
		"--label", filepath.Join("a", filename),
		"--label", filepath.Join("b", filename),
		origFile, fmtFile,
	)
	out, err := cmd.Output()
	if err != nil {
		// diff は差分がある場合 exit code 1 を返す
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return out, nil
		}
		// diffコマンドが見つからない場合など
		return fallbackDiff(filename, original, formatted), nil
	}
	return out, nil
}

// fallbackDiff はdiffコマンドが使えない場合のフォールバック出力
func fallbackDiff(filename string, original, formatted []byte) []byte {
	return []byte("--- " + filepath.Join("a", filename) + "\n" +
		"+++ " + filepath.Join("b", filename) + "\n" +
		"(diff command not available; showing formatted output)\n" +
		string(formatted))
}
