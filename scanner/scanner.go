package scanner

import (
	"fmt"
	"strconv"

	"github.com/zhiruchen/json-parse/token"
)

// Scanner json scanner
type Scanner struct {
	source   string
	runes    []rune
	tokens   []*token.Token
	start    int
	current  int
	line     int
	keywords map[string]token.Type

	errFunc ErrFunc
}

type ErrFunc func(line int, msg string)

// NewScanner a new json scanner
func NewScanner(source string, errFunc ErrFunc) *Scanner {
	return &Scanner{
		source: source,
		runes:  []rune(source),
		tokens: []*token.Token{},
		line:   1,
		keywords: map[string]token.Type{
			"false": token.False,
			"null":  token.Null,
			"true":  token.True,
		},
		errFunc: errFunc,
	}
}

// ScanTokens 返回扫描到的token列表
func (scan *Scanner) ScanTokens() []*token.Token {
	for !scan.isAtEnd() {
		scan.start = scan.current
		scan.scanToken()
	}
	scan.tokens = append(
		scan.tokens,
		&token.Token{
			TokenType: token.Eof,
			Lexeme:    "",
			Literal:   nil,
			Line:      scan.line,
		},
	)
	return scan.tokens
}

func (scan *Scanner) PrintTokens() {
	for _, t := range scan.tokens {
		fmt.Printf("%s\n", t.ToString())
	}
}

func (scan *Scanner) isAtEnd() bool {
	return scan.current >= len(scan.runes)
}

func (scan *Scanner) scanToken() {
	c := scan.advance()
	switch c {
	case '{':
		scan.addToken(token.LeftBrace, nil)
	case '}':
		scan.addToken(token.RightBrace, nil)
	case '[':
		scan.addToken(token.OpenBracket, nil)
	case ']':
		scan.addToken(token.CloseBracket, nil)
	case ':':
		scan.addToken(token.Colon, nil)
	case ',':
		scan.addToken(token.Comma, nil)
	case ' ', '\r', '\t': // 自动 break
	case '\n':
		scan.line++
	case '"':
		scan.getStr()
	default:
		if isDigits(c) {
			scan.getNumber()
		} else if isAlpha(c) {
			scan.getIdentifier()
		} else {
			scan.errFunc(scan.line, "Unexpected token!")
		}
	}
}

func (scan *Scanner) advance() rune {
	scan.current++
	return scan.runes[scan.current-1]
}

func (scan *Scanner) addToken(tokenType token.Type, literal interface{}) {
	text := string(scan.runes[scan.start:scan.current])
	scan.tokens = append(scan.tokens, &token.Token{TokenType: tokenType, Lexeme: text, Literal: literal, Line: scan.line})
}

func (scan *Scanner) match(expected rune) bool {
	if scan.isAtEnd() {
		return false
	}

	if scan.runes[scan.current] != expected {
		return false
	}

	scan.current++
	return true
}

func (scan *Scanner) peek() rune {
	if scan.current >= len(scan.runes) {
		return '\000' // https://stackoverflow.com/questions/38007361/is-there-anyway-to-create-null-terminated-string-in-go
	}
	return scan.runes[scan.current]
}

func (scan *Scanner) peekNext() rune {
	if (scan.current + 1) >= len(scan.runes) {
		return '\000'
	}
	return scan.runes[scan.current+1]
}

func (scan *Scanner) getStr() {
	for scan.peek() != '"' && !scan.isAtEnd() {
		if scan.peek() == '\n' {
			scan.line++
		}
		scan.advance()
	}
	if scan.isAtEnd() {
		scan.errFunc(scan.line, "Unterminated string")
		return
	}

	scan.advance()

	value := string(scan.runes[scan.start+1 : scan.current-1])
	scan.addToken(token.String, value)
}

func (scan *Scanner) getNumber() {
	for isDigits(scan.peek()) {
		scan.advance()
	}

	if scan.peek() == '.' && isDigits(scan.peekNext()) {
		scan.advance()

		for isDigits(scan.peek()) {
			scan.advance()
		}
	}

	text := string(scan.runes[scan.start:scan.current])
	number, _ := strconv.ParseFloat(text, 64)
	scan.addToken(token.Number, number)
}

func (scan *Scanner) getIdentifier() {
	for isAlphaNumberic(scan.peek()) {
		scan.advance()
	}

	text := string(scan.runes[scan.start:scan.current])
	tokenType, ok := scan.keywords[text]
	if !ok {
		scan.errFunc(scan.line, "unexpected value!")
		return
	}

	var literal interface{}
	switch tokenType {
	case token.Null:
		literal = "null"
	case token.True:
		literal = true
	case token.False:
		literal = false
	}

	scan.addToken(tokenType, literal)
}

func isDigits(c rune) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isAlphaNumberic(c rune) bool {
	return isAlpha(c) || isDigits(c)
}
