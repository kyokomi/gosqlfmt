package formatter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kyokomi/gosqlfmt/formatter"
)

func TestFormatSource(t *testing.T) {
	tests := []struct {
		name      string
		inputFile string
		wantFile  string
	}{
		{
			name:      "基本的なSQLフォーマット",
			inputFile: "simple.go",
			wantFile:  "simple.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath := filepath.Join("..", "testdata", "input", tt.inputFile)
			wantPath := filepath.Join("..", "testdata", "golden", tt.wantFile)

			input, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("入力ファイル読み込みエラー: %v", err)
			}

			want, err := os.ReadFile(wantPath)
			if err != nil {
				t.Fatalf("期待値ファイル読み込みエラー: %v", err)
			}

			got, err := formatter.FormatSource(input)
			if err != nil {
				t.Fatalf("FormatSource エラー: %v", err)
			}

			if string(got) != string(want) {
				t.Errorf("FormatSource の結果が期待値と異なる\n--- got ---\n%s\n--- want ---\n%s", string(got), string(want))
			}
		})
	}
}

func TestFormatSource_NonGoCode(t *testing.T) {
	// 不正なGoコードの場合はエラーを返す
	src := []byte("this is not valid go code")
	_, err := formatter.FormatSource(src)
	if err == nil {
		t.Error("不正なGoコードの場合はエラーを返すべき")
	}
}

func TestFormatSource_NoSQL(t *testing.T) {
	// SQLを含まないGoコードは変更されない
	src := []byte(`package main

func main() {
	msg := ` + "`Hello, World!`" + `
	_ = msg
}
`)
	got, err := formatter.FormatSource(src)
	if err != nil {
		t.Fatalf("FormatSource エラー: %v", err)
	}
	if string(got) != string(src) {
		t.Errorf("SQLを含まないコードは変更されるべきでない\n--- got ---\n%s\n--- want ---\n%s", string(got), string(src))
	}
}

func TestFormatSource_SQLDetection(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		shouldChange bool
	}{
		{
			name:       "SELECTで始まるSQL",
			src:        "package main\n\nvar q = `select * from users where status = ? and role = 'admin' order by created_at desc limit 10`\n",
			shouldChange: true,
		},
		{
			name:       "INSERTで始まるSQL",
			src:        "package main\n\nvar q = `insert into users (id, name, email, status) values (?, ?, ?, 'active')`\n",
			shouldChange: true,
		},
		{
			name:       "UPDATEで始まるSQL",
			src:        "package main\n\nvar q = `update users set name = ?, email = ?, updated_at = NOW() where id = ? and status = 'active'`\n",
			shouldChange: true,
		},
		{
			name:       "DELETEで始まるSQL",
			src:        "package main\n\nvar q = `delete from users where id = ? and status = 'inactive' and created_at < '2020-01-01'`\n",
			shouldChange: true,
		},
		{
			name:       "通常の文字列はSQLと判定されない",
			src:        "package main\n\nvar s = `This is just a regular string`\n",
			shouldChange: false,
		},
		{
			name:       "JSONはSQLと判定されない",
			src:        "package main\n\nvar j = `{\"key\": \"value\"}`\n",
			shouldChange: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatter.FormatSource([]byte(tt.src))
			if err != nil {
				t.Fatalf("FormatSource エラー: %v", err)
			}
			changed := string(got) != tt.src
			if changed != tt.shouldChange {
				t.Errorf("SQL検出結果が期待と異なる: changed=%v, shouldChange=%v\n--- input ---\n%s\n--- got ---\n%s", changed, tt.shouldChange, tt.src, string(got))
			}
		})
	}
}
