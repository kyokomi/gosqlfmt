package sqlfmt

import "strings"

// TokenType はSQLトークンの種類
type TokenType int

const (
	TokenKeyword     TokenType = iota // SQLキーワード
	TokenIdentifier                   // 識別子（テーブル名、カラム名等）
	TokenNumber                       // 数値リテラル
	TokenString                       // 文字列リテラル ('...')
	TokenOperator                     // 演算子 (=, <, >, !=, <=, >=, <>)
	TokenComma                        // ,
	TokenLParen                       // (
	TokenRParen                       // )
	TokenPlaceholder                  // ?, :name, $1
	TokenWhitespace                   // 空白
)

// Token はSQLの1トークン
type Token struct {
	Type    TokenType
	Literal string
}

// keywords はSQLキーワードの一覧（大文字）
var keywords = map[string]bool{
	"SELECT": true, "FROM": true, "WHERE": true,
	"JOIN": true, "INNER": true, "LEFT": true, "RIGHT": true, "OUTER": true, "CROSS": true,
	"ON": true, "AND": true, "OR": true,
	"GROUP": true, "ORDER": true, "BY": true,
	"HAVING": true, "LIMIT": true, "OFFSET": true,
	"INSERT": true, "INTO": true, "VALUES": true,
	"UPDATE": true, "SET": true, "DELETE": true,
	"AS": true, "IN": true, "NOT": true, "NULL": true, "IS": true,
	"BETWEEN": true, "LIKE": true, "EXISTS": true,
	"CASE": true, "WHEN": true, "THEN": true, "ELSE": true, "END": true,
	"UNION": true, "ALL": true, "DISTINCT": true,
	"ASC": true, "DESC": true,
}

// clauseKeywords は主要句キーワード（改行対象）
var clauseKeywords = map[string]bool{
	"SELECT": true, "FROM": true, "WHERE": true,
	"JOIN": true, "INNER JOIN": true, "LEFT JOIN": true, "RIGHT JOIN": true,
	"OUTER JOIN": true, "CROSS JOIN": true,
	"ON": true,
	"GROUP BY": true, "ORDER BY": true, "HAVING": true,
	"LIMIT": true, "OFFSET": true,
	"INSERT INTO": true, "VALUES": true,
	"UPDATE": true, "SET": true,
	"DELETE FROM": true,
	"UNION": true, "UNION ALL": true,
}

// compoundPairs は2語キーワードの1語目→2語目のマッピング
var compoundPairs = map[string]string{
	"INNER":  "JOIN",
	"LEFT":   "JOIN",
	"RIGHT":  "JOIN",
	"OUTER":  "JOIN",
	"CROSS":  "JOIN",
	"GROUP":  "BY",
	"ORDER":  "BY",
	"INSERT": "INTO",
	"DELETE": "FROM",
	"UNION":  "ALL",
}

// isClauseKeyword は主要句キーワードかどうか判定する
func isClauseKeyword(keyword string) bool {
	return clauseKeywords[strings.ToUpper(keyword)]
}
