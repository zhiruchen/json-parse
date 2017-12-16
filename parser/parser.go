package parser

import (
	"bytes"
	"fmt"

	"github.com/zhiruchen/json-parse/token"
)

var blankSpace = "    "

type ErrFunc func(tokenName string, msg string)

type Parser struct {
	tokens  []*token.Token
	current int
	errFunc ErrFunc
}

type JSONObject map[string]interface{}

func (obj JSONObject) Represent() string {
	if obj == nil {
		return "{}"
	}

	var count = len(obj)
	buf := bytes.NewBufferString("{")
	for k, v := range obj {
		buf.WriteString("\n" + blankSpace)
		buf.WriteString(k + ": ")

		switch v.(type) {
		case string:
			buf.WriteString(v.(string))
		case float64, bool:
			buf.WriteString(fmt.Sprintf("%v", v))
		case nil:
			buf.WriteString("null")
		case JSONObject:
			buf.WriteString("\n")
			vv := v.(JSONObject)
			buf.WriteString(vv.Represent())
		case JSONArray:
			vv := v.(JSONArray)
			buf.WriteString(vv.Represent())
		}

		if count > 1 {
			buf.WriteString(",")
		}
		count--
	}

	buf.WriteString("\n" + blankSpace + "}")
	return buf.String()
}

type pair struct {
	key string
	val interface{}
}

type JSONArray []interface{}

func (array JSONArray) Represent() string {
	if array == nil {
		return "[]"
	}

	var count = len(array)
	buf := bytes.NewBufferString("[")

	for _, v := range array {
		buf.WriteString("\n" + blankSpace)

		switch v.(type) {
		case string, float64, bool:
			buf.WriteString(fmt.Sprintf("%v", v))
		case nil:
			buf.WriteString("null")
		case JSONObject:
			buf.WriteString("\n")
			buf.WriteString(blankSpace)
			vv := v.(JSONObject)
			buf.WriteString(vv.Represent())
		case JSONArray:
			vv := v.(JSONArray)
			buf.WriteString(vv.Represent())
		}

		if count > 1 {
			buf.WriteString(",")
		}

		count--
	}

	buf.WriteString("\n" + "]")
	return fmt.Sprintf(buf.String())
}

func NewParser(tokens []*token.Token, errFunc ErrFunc) *Parser {
	return &Parser{tokens: tokens, errFunc: errFunc}
}

func (p *Parser) Parse() JSONer {

	if p.check(token.LeftBrace) {
		return p.object()
	}

	if p.check(token.OpenBracket) {
		return p.array()
	}

	p.errFunc(p.peek().ToString(), "unexpected token")
	return nil
}

func (p *Parser) object() JSONObject {
	p.consume(token.LeftBrace, "expect `{`")
	obj := make(JSONObject)

	if !p.check(token.RightBrace) {
		p.members(obj)
	}

	p.consume(token.RightBrace, fmt.Sprintf("expect `}` after %v", p.peek().Lexeme))

	return obj
}

func (p *Parser) members(obj JSONObject) {
	pair := p.pair()
	obj[pair.key] = pair.val

	for p.check(token.Comma) {
		p.consume(token.Comma, fmt.Sprintf("expect `,` after %v", pair.val))

		pair := p.pair()
		obj[pair.key] = pair.val
	}
}

func (p *Parser) pair() *pair {
	key := p.consume(token.String, "expect key")
	p.consume(token.Colon, fmt.Sprintf("expect `:` after key: %s", key))

	pr := &pair{key: key.Literal.(string)}
	pr.val = p.getValue()

	return pr
}

func (p *Parser) getValue() interface{} {
	if p.check(token.String) {
		val := p.consume(token.String, "expect string")
		return val.Literal.(string)
	}

	if p.check(token.Number) {
		val := p.consume(token.Number, "expect number")
		return val.Literal.(float64)
	}

	if p.check(token.True) {
		val := p.consume(token.True, "expect true")
		return val.Literal.(bool)
	}

	if p.check(token.False) {
		val := p.consume(token.False, "expect false")
		return val.Literal.(bool)
	}

	if p.check(token.Null) {
		p.consume(token.Null, "expect null")
		return nil
	}

	if p.check(token.LeftBrace) {
		return p.object()
	}

	if p.check(token.OpenBracket) {
		return p.array()
	}

	p.errFunc(p.peek().ToString(), "unsupported value type")
	return nil
}

func (p *Parser) array() JSONArray {
	p.consume(token.OpenBracket, "expect `[`")
	array := p.elements()
	p.consume(token.CloseBracket, "expect `]`")

	return array
}

func (p *Parser) elements() JSONArray {
	val := p.getValue()
	array := JSONArray{val}

	for p.check(token.Comma) {
		p.consume(token.Comma, fmt.Sprintf("expect `,` after %v", val))

		array = append(array, p.getValue())
	}

	return array
}

func (p *Parser) consume(t token.Type, msg string) *token.Token {
	if p.check(t) {
		return p.advance()
	}

	p.errFunc("", msg)
	return nil
}

func (p *Parser) match(tokenTypes ...token.Type) bool {
	for _, t := range tokenTypes {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(t token.Type) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().TokenType == t
}

func (p *Parser) peek() *token.Token {
	return p.tokens[p.current]
}

func (p *Parser) advance() *token.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().TokenType == token.Eof
}

func (p *Parser) previous() *token.Token {
	return p.tokens[p.current-1]
}
