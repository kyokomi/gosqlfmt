package sqlfmt

import "strings"

// DefaultLineWidth は複数行展開の閾値（デフォルト80文字）
const DefaultLineWidth = 80

// Normalize はキーワードの大文字化と無駄なスペースの正規化のみ行う（1行のまま）。
func Normalize(sql string) string {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return ""
	}

	lexer := NewLexer(trimmed)
	tokens := lexer.Tokenize()

	var result strings.Builder
	for i, tok := range tokens {
		if tok.Type == TokenWhitespace {
			if i > 0 && i < len(tokens)-1 {
				result.WriteString(" ")
			}
			continue
		}
		result.WriteString(tok.Literal)
	}
	return result.String()
}

// Format はSQL文字列をフォーマットする。
// 閾値を超える場合は複数行に展開し、閾値以下の場合はキーワード大文字化とスペース正規化のみ行う。
func Format(sql string) string {
	return FormatWithWidth(sql, DefaultLineWidth)
}

// FormatWithWidth は指定した閾値でSQL文字列をフォーマットする。
func FormatWithWidth(sql string, lineWidth int) string {
	normalized := Normalize(sql)
	if normalized == "" {
		return ""
	}
	if lineWidth > 0 && len(normalized) <= lineWidth {
		return normalized
	}
	return formatMultiLine(normalized)
}

// formatMultiLine はSQL文字列を複数行にフォーマットする
func formatMultiLine(sql string) string {
	lexer := NewLexer(sql)
	tokens := lexer.Tokenize()
	tokens = mergeCompoundKeywords(tokens)
	clauses := splitIntoClauses(tokens)

	var result strings.Builder
	result.WriteString("\n")

	for i, c := range clauses {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(c.keyword)

		if len(c.body) == 0 {
			continue
		}

		switch c.keyword {
		case "SELECT", "SET":
			result.WriteString(formatCommaSeparated(c.body))
		case "WHERE", "ON", "HAVING":
			result.WriteString(formatConditions(c.body))
		default:
			result.WriteString(formatDefaultBody(c.body))
		}
	}

	result.WriteString("\n")
	return result.String()
}

// mergeCompoundKeywords は2語キーワード（INNER JOIN, GROUP BY等）を1トークンに結合する
func mergeCompoundKeywords(tokens []Token) []Token {
	var result []Token
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok.Type == TokenKeyword {
			if second, ok := compoundPairs[tok.Literal]; ok {
				j := i + 1
				for j < len(tokens) && tokens[j].Type == TokenWhitespace {
					j++
				}
				if j < len(tokens) && tokens[j].Type == TokenKeyword && tokens[j].Literal == second {
					result = append(result, Token{TokenKeyword, tok.Literal + " " + tokens[j].Literal})
					i = j
					continue
				}
			}
		}
		result = append(result, tokens[i])
	}
	return result
}

type sqlClause struct {
	keyword string
	body    []Token
}

// splitIntoClauses はトークン列を主要句ごとに分割する
func splitIntoClauses(tokens []Token) []sqlClause {
	var clauses []sqlClause
	var currentKeyword string
	var currentBody []Token
	parenDepth := 0

	for _, tok := range tokens {
		switch tok.Type {
		case TokenLParen:
			parenDepth++
		case TokenRParen:
			parenDepth--
		}

		if parenDepth == 0 && tok.Type == TokenKeyword && isClauseKeyword(tok.Literal) {
			if currentKeyword != "" {
				clauses = append(clauses, sqlClause{keyword: currentKeyword, body: currentBody})
			}
			currentKeyword = tok.Literal
			currentBody = nil
		} else {
			currentBody = append(currentBody, tok)
		}
	}

	if currentKeyword != "" {
		clauses = append(clauses, sqlClause{keyword: currentKeyword, body: currentBody})
	}

	return clauses
}

// formatCommaSeparated はカンマで区切って改行+インデントする（SELECT句、SET句用）
func formatCommaSeparated(body []Token) string {
	parts := splitByComma(body)
	var result strings.Builder
	for i, part := range parts {
		if i == 0 {
			result.WriteString("\n  ")
		} else {
			result.WriteString(",\n  ")
		}
		result.WriteString(tokensToStringTrimmed(part))
	}
	return result.String()
}

type conditionPart struct {
	connector string // "", "AND", "OR"
	tokens    []Token
}

// formatConditions はAND/ORで改行+インデントする（WHERE句、ON句、HAVING句用）
func formatConditions(body []Token) string {
	parts := splitByConditions(body)
	var result strings.Builder
	for _, part := range parts {
		result.WriteString("\n  ")
		if part.connector != "" {
			result.WriteString(part.connector + " ")
		}
		result.WriteString(tokensToStringTrimmed(part.tokens))
	}
	return result.String()
}

// formatDefaultBody はデフォルトのbody整形（改行+インデント）
func formatDefaultBody(body []Token) string {
	trimmed := tokensToStringTrimmed(body)
	if trimmed == "" {
		return ""
	}
	return "\n  " + trimmed
}

// splitByComma は括弧外のカンマでトークン列を分割する
func splitByComma(body []Token) [][]Token {
	var parts [][]Token
	var current []Token
	parenDepth := 0

	for _, tok := range body {
		switch tok.Type {
		case TokenLParen:
			parenDepth++
		case TokenRParen:
			parenDepth--
		}
		if parenDepth == 0 && tok.Type == TokenComma {
			parts = append(parts, current)
			current = nil
			continue
		}
		current = append(current, tok)
	}

	if len(current) > 0 {
		parts = append(parts, current)
	}
	return parts
}

// splitByConditions は括弧外のAND/ORでトークン列を分割する
func splitByConditions(body []Token) []conditionPart {
	var parts []conditionPart
	var current []Token
	connector := ""
	parenDepth := 0

	for _, tok := range body {
		if tok.Type == TokenLParen {
			parenDepth++
		} else if tok.Type == TokenRParen {
			parenDepth--
		}
		if parenDepth == 0 && tok.Type == TokenKeyword && (tok.Literal == "AND" || tok.Literal == "OR") {
			parts = append(parts, conditionPart{connector: connector, tokens: current})
			connector = tok.Literal
			current = nil
			continue
		}
		current = append(current, tok)
	}

	if len(current) > 0 {
		parts = append(parts, conditionPart{connector: connector, tokens: current})
	}
	return parts
}

// tokensToStringTrimmed はトークン列を文字列に変換する（前後の空白除去、中間の空白は1スペースに）
func tokensToStringTrimmed(tokens []Token) string {
	start := 0
	for start < len(tokens) && tokens[start].Type == TokenWhitespace {
		start++
	}
	end := len(tokens)
	for end > start && tokens[end-1].Type == TokenWhitespace {
		end--
	}

	var result strings.Builder
	for i := start; i < end; i++ {
		tok := tokens[i]
		if tok.Type == TokenWhitespace {
			result.WriteString(" ")
		} else {
			result.WriteString(tok.Literal)
		}
	}
	return result.String()
}
