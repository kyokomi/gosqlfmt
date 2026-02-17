package formatter

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/kyokomi/gosqlfmt/sqlfmt"
)

// FormatSource はGoソースコード内のSQL文字列リテラルをフォーマットする。
func FormatSource(src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// 置換対象を収集（位置と元文字列→新文字列のペア）
	type replacement struct {
		start int
		end   int
		old   string
		new   string
	}
	var replacements []replacement

	ast.Inspect(f, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}

		// バッククォート文字列のみ対象
		if !strings.HasPrefix(lit.Value, "`") || !strings.HasSuffix(lit.Value, "`") {
			return true
		}

		raw := lit.Value[1 : len(lit.Value)-1] // バッククォートを除去

		if !isSQL(raw) {
			return true
		}

		formatted := sqlfmt.Format(raw)
		if formatted == raw {
			return true
		}

		newLit := "`" + formatted + "`"
		if newLit == lit.Value {
			return true
		}

		offset := fset.Position(lit.Pos()).Offset
		replacements = append(replacements, replacement{
			start: offset,
			end:   offset + len(lit.Value),
			old:   lit.Value,
			new:   newLit,
		})

		return true
	})

	if len(replacements) == 0 {
		return src, nil
	}

	// 後ろから置換して位置がずれないようにする
	result := make([]byte, len(src))
	copy(result, src)
	for i := len(replacements) - 1; i >= 0; i-- {
		r := replacements[i]
		result = append(result[:r.start], append([]byte(r.new), result[r.end:]...)...)
	}

	return result, nil
}

// isSQL はバッククォート内の文字列がSQLかどうか判定する
func isSQL(s string) bool {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return false
	}
	upper := strings.ToUpper(trimmed)
	sqlPrefixes := []string{"SELECT ", "INSERT ", "UPDATE ", "DELETE ", "CREATE ", "ALTER ", "DROP "}
	for _, prefix := range sqlPrefixes {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}
