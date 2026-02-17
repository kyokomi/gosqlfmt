package sqlfmt

import (
	"strings"
	"unicode"
)

// Lexer はSQL文字列の字句解析器
type Lexer struct {
	input []rune
	pos   int
}

// NewLexer は新しいLexerを作成する
func NewLexer(input string) *Lexer {
	return &Lexer{input: []rune(input)}
}

// Tokenize はSQL文字列をトークン列に分解する
func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		switch {
		case unicode.IsSpace(ch):
			tokens = append(tokens, l.readWhitespace())
		case ch == '\'':
			tokens = append(tokens, l.readStringLiteral())
		case ch == '?':
			tokens = append(tokens, Token{TokenPlaceholder, "?"})
			l.pos++
		case ch == ':' && l.pos+1 < len(l.input) && unicode.IsLetter(l.input[l.pos+1]):
			tokens = append(tokens, l.readNamedPlaceholder())
		case ch == '$' && l.pos+1 < len(l.input) && unicode.IsDigit(l.input[l.pos+1]):
			tokens = append(tokens, l.readPositionalPlaceholder())
		case unicode.IsLetter(ch) || ch == '_':
			tokens = append(tokens, l.readIdentifierOrKeyword())
		case unicode.IsDigit(ch):
			tokens = append(tokens, l.readNumber())
		case ch == ',':
			tokens = append(tokens, Token{TokenComma, ","})
			l.pos++
		case ch == '(':
			tokens = append(tokens, Token{TokenLParen, "("})
			l.pos++
		case ch == ')':
			tokens = append(tokens, Token{TokenRParen, ")"})
			l.pos++
		case ch == '*':
			tokens = append(tokens, Token{TokenIdentifier, "*"})
			l.pos++
		default:
			tokens = append(tokens, l.readOperator())
		}
	}
	return tokens
}

func (l *Lexer) readWhitespace() Token {
	start := l.pos
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
	return Token{TokenWhitespace, string(l.input[start:l.pos])}
}

func (l *Lexer) readStringLiteral() Token {
	start := l.pos
	l.pos++ // skip opening '
	for l.pos < len(l.input) {
		if l.input[l.pos] == '\'' {
			if l.pos+1 < len(l.input) && l.input[l.pos+1] == '\'' {
				l.pos += 2 // エスケープされた ''
				continue
			}
			l.pos++ // closing '
			break
		}
		l.pos++
	}
	return Token{TokenString, string(l.input[start:l.pos])}
}

func (l *Lexer) readNamedPlaceholder() Token {
	start := l.pos
	l.pos++ // skip ':'
	for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
		l.pos++
	}
	return Token{TokenPlaceholder, string(l.input[start:l.pos])}
}

func (l *Lexer) readPositionalPlaceholder() Token {
	start := l.pos
	l.pos++ // skip '$'
	for l.pos < len(l.input) && unicode.IsDigit(l.input[l.pos]) {
		l.pos++
	}
	return Token{TokenPlaceholder, string(l.input[start:l.pos])}
}

func (l *Lexer) readIdentifierOrKeyword() Token {
	start := l.pos
	for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
		l.pos++
	}
	// ドット付き識別子（テーブル名.カラム名）
	for l.pos < len(l.input) && l.input[l.pos] == '.' {
		next := l.pos + 1
		if next < len(l.input) && (unicode.IsLetter(l.input[next]) || l.input[next] == '_') {
			l.pos++ // skip '.'
			for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
				l.pos++
			}
		} else {
			break
		}
	}
	word := string(l.input[start:l.pos])
	upper := strings.ToUpper(word)
	if keywords[upper] {
		return Token{TokenKeyword, upper}
	}
	return Token{TokenIdentifier, word}
}

func (l *Lexer) readNumber() Token {
	start := l.pos
	for l.pos < len(l.input) && (unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '.') {
		l.pos++
	}
	return Token{TokenNumber, string(l.input[start:l.pos])}
}

func (l *Lexer) readOperator() Token {
	start := l.pos
	ch := l.input[l.pos]
	l.pos++
	if l.pos < len(l.input) {
		next := l.input[l.pos]
		twoChar := string([]rune{ch, next})
		if twoChar == "!=" || twoChar == "<>" || twoChar == "<=" || twoChar == ">=" {
			l.pos++
			return Token{TokenOperator, twoChar}
		}
	}
	return Token{TokenOperator, string(l.input[start:l.pos])}
}
