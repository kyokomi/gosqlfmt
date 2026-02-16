package sqlfmt

// DefaultLineWidth は複数行展開の閾値（デフォルト80文字）
const DefaultLineWidth = 80

// Format はSQL文字列をフォーマットする。
// 閾値を超える場合は複数行に展開し、閾値以下の場合はキーワード大文字化とスペース正規化のみ行う。
func Format(sql string) string {
	return sql
}

// FormatWithWidth は指定した閾値でSQL文字列をフォーマットする。
func FormatWithWidth(sql string, lineWidth int) string {
	return sql
}

// Normalize はキーワードの大文字化と無駄なスペースの正規化のみ行う（1行のまま）。
func Normalize(sql string) string {
	return sql
}
