package main_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "gosqlfmt-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	binaryPath = filepath.Join(dir, "gosqlfmt")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("バイナリのビルドに失敗: " + err.Error())
	}

	os.Exit(m.Run())
}

func TestCLI_FileToStdout(t *testing.T) {
	inputPath := filepath.Join("..", "..", "testdata", "input", "simple.go")
	goldenPath := filepath.Join("..", "..", "testdata", "golden", "simple.go")

	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("goldenファイル読み込みエラー: %v", err)
	}

	cmd := exec.Command(binaryPath, inputPath)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("コマンド実行エラー: %v", err)
	}

	if string(out) != string(golden) {
		t.Errorf("stdout出力が期待値と異なる\n--- got ---\n%s\n--- want ---\n%s", string(out), string(golden))
	}
}

func TestCLI_WriteFlag(t *testing.T) {
	// テスト用の一時ファイルを作成
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")

	input := []byte("package main\n\nvar q = `select * from hoge where id = ?`\n")
	if err := os.WriteFile(tmpFile, input, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(binaryPath, "-w", tmpFile)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("コマンド実行エラー: %v\n%s", err, string(out))
	}

	got, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	want := "package main\n\nvar q = `SELECT * FROM hoge WHERE id = ?`\n"
	if string(got) != want {
		t.Errorf("-w フラグの結果が期待値と異なる\n--- got ---\n%s\n--- want ---\n%s", string(got), want)
	}
}

func TestCLI_ListFlag(t *testing.T) {
	t.Run("要フォーマットファイル", func(t *testing.T) {
		inputPath := filepath.Join("..", "..", "testdata", "input", "simple.go")
		cmd := exec.Command(binaryPath, "-l", inputPath)
		out, err := cmd.Output()
		if err == nil {
			t.Fatal("変更があるファイルに対して -l は非ゼロ終了コードを返すべき")
		}
		exitErr, ok := err.(*exec.ExitError)
		if !ok || exitErr.ExitCode() != 1 {
			t.Fatalf("終了コードが1であるべき: %v", err)
		}
		if !strings.Contains(string(out), inputPath) {
			t.Errorf("-l の出力にファイルパスが含まれるべき: %s", string(out))
		}
	})

	t.Run("フォーマット済みファイル", func(t *testing.T) {
		goldenPath := filepath.Join("..", "..", "testdata", "golden", "simple.go")
		cmd := exec.Command(binaryPath, "-l", goldenPath)
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("フォーマット済みファイルに対して -l はゼロ終了コードを返すべき: %v", err)
		}
		if len(strings.TrimSpace(string(out))) > 0 {
			t.Errorf("フォーマット済みファイルに対して -l は出力しないべき: %s", string(out))
		}
	})
}

func TestCLI_DiffFlag(t *testing.T) {
	inputPath := filepath.Join("..", "..", "testdata", "input", "simple.go")
	cmd := exec.Command(binaryPath, "-d", inputPath)
	out, err := cmd.Output()
	if err == nil {
		t.Fatal("変更があるファイルに対して -d は非ゼロ終了コードを返すべき")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 1 {
		t.Fatalf("終了コードが1であるべき: %v", err)
	}
	// unified diff 形式のヘッダーを確認
	if !strings.Contains(string(out), "---") || !strings.Contains(string(out), "+++") {
		t.Errorf("-d の出力がunified diff形式でない: %s", string(out))
	}
}

func TestCLI_Stdin(t *testing.T) {
	input := []byte("package main\n\nvar q = `select * from hoge where id = ?`\n")
	cmd := exec.Command(binaryPath)
	cmd.Stdin = bytes.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("コマンド実行エラー: %v", err)
	}

	want := "package main\n\nvar q = `SELECT * FROM hoge WHERE id = ?`\n"
	if string(out) != want {
		t.Errorf("stdin経由の出力が期待値と異なる\n--- got ---\n%s\n--- want ---\n%s", string(out), want)
	}
}

func TestCLI_Directory(t *testing.T) {
	// 一時ディレクトリにテスト用ファイルを作成
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "a.go")
	file2 := filepath.Join(tmpDir, "b.go")
	os.WriteFile(file1, []byte("package main\n\nvar q = `select * from hoge where id = ?`\n"), 0o644)
	os.WriteFile(file2, []byte("package main\n\nvar q = `SELECT * FROM hoge WHERE id = ?`\n"), 0o644)

	cmd := exec.Command(binaryPath, "-l", tmpDir)
	out, err := cmd.Output()

	// file1 は要フォーマットなので exit code 1
	if err == nil {
		t.Fatal("変更があるファイルがある場合は非ゼロ終了コードを返すべき")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 1 {
		t.Fatalf("終了コードが1であるべき: %v", err)
	}

	output := string(out)
	if !strings.Contains(output, "a.go") {
		t.Errorf("要フォーマットファイル a.go が一覧に含まれるべき: %s", output)
	}
	if strings.Contains(output, "b.go") {
		t.Errorf("フォーマット済みファイル b.go が一覧に含まれるべきでない: %s", output)
	}
}

func TestCLI_ExitCode(t *testing.T) {
	t.Run("変更なしの場合はexit0", func(t *testing.T) {
		goldenPath := filepath.Join("..", "..", "testdata", "golden", "simple.go")
		cmd := exec.Command(binaryPath, "-l", goldenPath)
		err := cmd.Run()
		if err != nil {
			t.Errorf("フォーマット済みファイルの場合は終了コード0であるべき: %v", err)
		}
	})

	t.Run("変更ありの場合はexit1", func(t *testing.T) {
		inputPath := filepath.Join("..", "..", "testdata", "input", "simple.go")
		cmd := exec.Command(binaryPath, "-l", inputPath)
		err := cmd.Run()
		if err == nil {
			t.Fatal("要フォーマットファイルの場合は非ゼロ終了コードを返すべき")
		}
		exitErr, ok := err.(*exec.ExitError)
		if !ok || exitErr.ExitCode() != 1 {
			t.Fatalf("終了コードが1であるべき: %v", err)
		}
	})
}
