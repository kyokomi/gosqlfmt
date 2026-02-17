package sqlfmt

import "strings"

// DefaultLineWidth は複数行展開の閾値（デフォルト80文字）
const DefaultLineWidth = 80

// maxSubqueryDepth はサブクエリの最大再帰深度
const maxSubqueryDepth = 5

// subqueryKeywords はサブクエリの先頭に来るキーワード
var subqueryKeywords = map[string]bool{
	"SELECT": true, "INSERT": true, "UPDATE": true, "DELETE": true,
}

// Normalize はキーワードの大文字化と無駄なスペースの正規化のみ行う（1行のまま）。
func Normalize(sql string) string {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return ""
	}

	lexer := NewLexer(trimmed)
	tokens := lexer.Tokenize()

	var result strings.Builder
	prevLineComment := false
	for i, tok := range tokens {
		if tok.Type == TokenWhitespace {
			if prevLineComment {
				result.WriteString("\n")
				prevLineComment = false
			} else if i > 0 && i < len(tokens)-1 {
				result.WriteString(" ")
			}
			continue
		}
		if tok.Type == TokenLineComment {
			if i > 0 && !prevLineComment {
				prev := tokens[i-1]
				if prev.Type != TokenWhitespace {
					result.WriteString(" ")
				}
			}
			result.WriteString(tok.Literal)
			prevLineComment = true
			continue
		}
		if tok.Type == TokenBlockComment {
			if i > 0 {
				prev := tokens[i-1]
				if prev.Type != TokenWhitespace {
					result.WriteString(" ")
				}
			}
			result.WriteString(tok.Literal)
			prevLineComment = false
			continue
		}
		prevLineComment = false
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
	return formatMultiLineWithIndent(sql, "", 0)
}

// formatMultiLineWithIndent はインデント付きでSQL文字列を複数行にフォーマットする
func formatMultiLineWithIndent(sql string, indent string, depth int) string {
	lexer := NewLexer(sql)
	tokens := lexer.Tokenize()
	tokens = mergeCompoundKeywords(tokens)
	clauses := splitIntoClauses(tokens)

	bodyIndent := indent + "  "

	var result strings.Builder
	result.WriteString("\n")

	for i, c := range clauses {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(indent + c.keyword)

		if len(c.body) == 0 {
			continue
		}

		switch c.keyword {
		case "SELECT", "SET":
			result.WriteString(formatCommaSeparated(c.body, bodyIndent, depth))
		case "WHERE", "ON", "HAVING":
			result.WriteString(formatConditions(c.body, bodyIndent, depth))
		default:
			result.WriteString(formatDefaultBody(c.body, bodyIndent, depth))
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
	caseDepth := 0

	for _, tok := range tokens {
		switch tok.Type {
		case TokenLParen:
			parenDepth++
		case TokenRParen:
			parenDepth--
		}

		if parenDepth == 0 && tok.Type == TokenKeyword {
			if tok.Literal == "CASE" {
				caseDepth++
			} else if tok.Literal == "END" && caseDepth > 0 {
				caseDepth--
			}
		}

		if parenDepth == 0 && caseDepth == 0 && tok.Type == TokenKeyword && isClauseKeyword(tok.Literal) {
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
func formatCommaSeparated(body []Token, indent string, depth int) string {
	parts := splitByComma(body)
	var result strings.Builder
	for i, part := range parts {
		if i == 0 {
			result.WriteString("\n" + indent)
		} else {
			result.WriteString(",\n" + indent)
		}
		if containsCase(part) {
			result.WriteString(formatCaseExpression(part, indent))
		} else {
			result.WriteString(formatTokensWithSubquery(part, indent, depth))
		}
	}
	return result.String()
}

type conditionPart struct {
	connector string // "", "AND", "OR"
	tokens    []Token
}

// formatConditions はAND/ORで改行+インデントする（WHERE句、ON句、HAVING句用）
func formatConditions(body []Token, indent string, depth int) string {
	parts := splitByConditions(body)
	var result strings.Builder
	for _, part := range parts {
		result.WriteString("\n" + indent)
		if part.connector != "" {
			result.WriteString(part.connector + " ")
		}
		if containsCase(part.tokens) {
			result.WriteString(formatCaseExpression(part.tokens, indent))
		} else {
			result.WriteString(formatTokensWithSubquery(part.tokens, indent, depth))
		}
	}
	return result.String()
}

// formatDefaultBody はデフォルトのbody整形（改行+インデント）
func formatDefaultBody(body []Token, indent string, depth int) string {
	formatted := formatTokensWithSubquery(body, indent, depth)
	if formatted == "" {
		return ""
	}
	return "\n" + indent + formatted
}

// formatTokensWithSubquery はトークン列内のサブクエリを検出して再帰整形する
func formatTokensWithSubquery(tokens []Token, indent string, depth int) string {
	trimmed := trimWhitespaceTokens(tokens)
	if len(trimmed) == 0 {
		return ""
	}

	if depth >= maxSubqueryDepth {
		return tokensToStringTrimmed(tokens)
	}

	// サブクエリを含む括弧があるか検出
	var result strings.Builder
	i := 0
	needSpace := false

	for i < len(trimmed) {
		tok := trimmed[i]

		if tok.Type == TokenWhitespace {
			needSpace = true
			i++
			continue
		}

		if tok.Type == TokenLParen {
			// 対応する閉じ括弧を見つける
			parenStart := i
			parenDepth := 1
			j := i + 1
			for j < len(trimmed) && parenDepth > 0 {
				switch trimmed[j].Type {
				case TokenLParen:
					parenDepth++
				case TokenRParen:
					parenDepth--
				}
				j++
			}
			// trimmed[parenStart+1 : j-1] が括弧内のトークン
			inner := trimmed[parenStart+1 : j-1]

			if containsSubquery(inner) {
				// サブクエリとして再帰整形
				subIndent := indent + "  "
				innerSQL := tokensToStringTrimmed(inner)
				formatted := formatMultiLineWithIndent(innerSQL, subIndent, depth+1)

				if needSpace {
					result.WriteString(" ")
					needSpace = false
				}
				result.WriteString("(")
				result.WriteString(formatted)
				result.WriteString(indent + ")")
			} else {
				// 通常の括弧
				if needSpace {
					result.WriteString(" ")
					needSpace = false
				}
				result.WriteString(tokensToString(trimmed[parenStart:j]))
			}
			i = j
			continue
		}

		if needSpace {
			result.WriteString(" ")
			needSpace = false
		}
		result.WriteString(tok.Literal)
		i++
	}

	return result.String()
}

// containsSubquery は括弧内のトークンにサブクエリのキーワードがあるか判定する
func containsSubquery(tokens []Token) bool {
	for _, tok := range tokens {
		if tok.Type == TokenWhitespace {
			continue
		}
		return tok.Type == TokenKeyword && subqueryKeywords[tok.Literal]
	}
	return false
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

// splitByConditions は括弧外のAND/ORでトークン列を分割する。
// BETWEEN...AND の AND では分割しない。
func splitByConditions(body []Token) []conditionPart {
	var parts []conditionPart
	var current []Token
	connector := ""
	parenDepth := 0
	afterBetween := false

	for _, tok := range body {
		switch tok.Type {
		case TokenLParen:
			parenDepth++
		case TokenRParen:
			parenDepth--
		}
		if parenDepth == 0 && tok.Type == TokenKeyword {
			if tok.Literal == "BETWEEN" {
				afterBetween = true
			} else if tok.Literal == "AND" && afterBetween {
				afterBetween = false
				current = append(current, tok)
				continue
			} else if tok.Literal == "AND" || tok.Literal == "OR" {
				parts = append(parts, conditionPart{connector: connector, tokens: current})
				connector = tok.Literal
				current = nil
				continue
			}
		}
		current = append(current, tok)
	}

	if len(current) > 0 {
		parts = append(parts, conditionPart{connector: connector, tokens: current})
	}
	return parts
}

// containsCase はトークン列にparenDepth==0のCASEキーワードがあるか判定する
func containsCase(tokens []Token) bool {
	parenDepth := 0
	for _, tok := range tokens {
		switch tok.Type {
		case TokenLParen:
			parenDepth++
		case TokenRParen:
			parenDepth--
		}
		if parenDepth == 0 && tok.Type == TokenKeyword && tok.Literal == "CASE" {
			return true
		}
	}
	return false
}

// formatCaseExpression はCASE式を含むトークン列を整形する。
// baseIndent は現在のインデント（通常 "  "）。
func formatCaseExpression(tokens []Token, baseIndent string) string {
	trimmed := trimWhitespaceTokens(tokens)
	if len(trimmed) == 0 {
		return ""
	}

	var result strings.Builder
	caseDepth := 0
	parenDepth := 0
	needSpace := false

	for _, tok := range trimmed {
		switch tok.Type {
		case TokenLParen:
			parenDepth++
		case TokenRParen:
			parenDepth--
		}

		if parenDepth > 0 || tok.Type == TokenWhitespace {
			if tok.Type == TokenWhitespace {
				needSpace = true
			} else {
				if needSpace {
					result.WriteString(" ")
					needSpace = false
				}
				result.WriteString(tok.Literal)
			}
			continue
		}

		if tok.Type == TokenKeyword {
			switch tok.Literal {
			case "CASE":
				caseDepth++
				if caseDepth == 1 {
					if needSpace {
						result.WriteString(" ")
						needSpace = false
					}
					result.WriteString("CASE")
					continue
				}
			case "WHEN":
				if caseDepth == 1 {
					needSpace = false
					result.WriteString("\n" + baseIndent + "  WHEN")
					continue
				}
			case "THEN":
				if caseDepth == 1 {
					if needSpace {
						result.WriteString(" ")
						needSpace = false
					}
					result.WriteString("THEN")
					continue
				}
			case "ELSE":
				if caseDepth == 1 {
					needSpace = false
					result.WriteString("\n" + baseIndent + "  ELSE")
					continue
				}
			case "END":
				if caseDepth == 1 {
					needSpace = false
					result.WriteString("\n" + baseIndent + "END")
					caseDepth--
					continue
				}
				if caseDepth > 1 {
					caseDepth--
				}
			}
		}

		if needSpace {
			result.WriteString(" ")
			needSpace = false
		}
		result.WriteString(tok.Literal)
	}
	return result.String()
}

// trimWhitespaceTokens はトークン列の前後の空白トークンを除去する
func trimWhitespaceTokens(tokens []Token) []Token {
	start := 0
	for start < len(tokens) && tokens[start].Type == TokenWhitespace {
		start++
	}
	end := len(tokens)
	for end > start && tokens[end-1].Type == TokenWhitespace {
		end--
	}
	return tokens[start:end]
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

// tokensToString はトークン列を文字列に変換する（空白は1スペースに）
func tokensToString(tokens []Token) string {
	var result strings.Builder
	for _, tok := range tokens {
		if tok.Type == TokenWhitespace {
			result.WriteString(" ")
		} else {
			result.WriteString(tok.Literal)
		}
	}
	return result.String()
}
