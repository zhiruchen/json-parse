package parse

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// JSONTokener json token
type JSONTokener struct {
	runepos           int64 // 当前字符位置
	runes             []rune
	eof               bool  // 是否结束
	index             int   // 当前读取index
	line              int   // 当前读取行
	previous          rune  // 前一个字符
	usePrevious       bool  // 前一个自符是否被请求
	runesPreviousLine int64 // 前一行读取的字符数
}

type JSONArray struct {
	objects []interface{}
}

type JSONObject struct {
	data map[string]interface{}
}

func NewJSONArray(x *JSONTokener) (*JSONArray, error) {
	array := &JSONArray{objects: []interface{}{}}

	if x.nextNoneEmptyRune() != '[' {
		return nil, errors.New("json array must start with [")
	}
	if x.nextNoneEmptyRune() != ']' {
		x.back()
		for {
			if x.nextNoneEmptyRune() == ',' {
				array.objects = append(array.objects, nil)
			} else {
				x.back()
				v, err := x.nextValue()
				if err != nil {
					return nil, err
				}
				array.objects = append(array.objects, v)
			}

			switch x.nextNoneEmptyRune() {
			case ',':
				if x.nextNoneEmptyRune() == ']' {
					return array, nil
				}
				x.back()
			case ']':
				return array, nil
			default:
				return nil, errors.New("expected a ',' or ']'")
			}

		}
	}
	return array, nil
}

func NewJSONObject(x *JSONTokener) (*JSONObject, error) {
	obj := &JSONObject{data: make(map[string]interface{})}

	var r rune
	var key string

	if x.nextNoneEmptyRune() != '{' {
		return nil, errors.New("json text must start with {")
	}

	for {
		r = x.nextNoneEmptyRune()
		switch r {
		case '\000':
			return nil, errors.New("json text must end with }")
		case '}':
			return obj, nil
		default:
			x.back()
			// log.Printf("index: %d, currentRune: %#U\n", x.index, x.runes[x.index])
			v, err := x.nextValue()
			if err != nil {
				return nil, err
			}
			key = fmt.Sprintf("%v", v)
		}

		r = x.nextNoneEmptyRune()
		if r != ':' {
			return nil, fmt.Errorf("expect : after key: %s", key)
		}

		v, err := x.nextValue()
		if err != nil {
			return nil, err
		}
		obj.data[key] = v

		switch x.nextNoneEmptyRune() {
		case ';', ',':
			if x.nextNoneEmptyRune() == '}' {
				return obj, nil
			}
			x.back()
		case '}':
			return obj, nil
		default:
			return nil, errors.New("expected a , or }")
		}

	}
}
func (obj *JSONObject) GetValue(path string) (interface{}, error) {
	if path == "" {
		return nil, errors.New("path is empty")
	}

	k := strings.Split(path, ".")
	v, ok := obj.data[k[0]]
	if !ok {
		return nil, fmt.Errorf("no key: %s", k[0])
	}
	if len(k) == 1 {
		return v, nil
	}

	v1, ok1 := v.(*JSONObject)
	if !ok1 {
		return nil, fmt.Errorf("%s's value is not json object", k[0])
	}
	return v1.GetValue(strings.Join(k[1:], "."))
}

// ToDo: 字符串转数字处理
func stringToValue(s string) (interface{}, error) {
	if s == "" {
		return s, nil
	}
	if s == "true" {
		return true, nil
	}
	if s == "false" {
		return false, nil
	}

	if s == "null" {
		return nil, nil
	}

	rs := []rune(s)
	if (rs[0] >= '0' && rs[0] <= '9') || rs[0] == '-' {
		v1, err1 := strconv.ParseFloat(s, 64)
		if err1 != nil {
			return nil, err1
		}
		return v1, nil
	}
	return s, nil
}

func NewJSONTokener(js string) (*JSONTokener, error) {
	jsonTokener := &JSONTokener{
		runepos:           1,
		eof:               false,
		index:             0,
		line:              0,
		previous:          0,
		usePrevious:       false,
		runesPreviousLine: 0,
	}
	jsonTokener.runes = []rune(js)
	// log.Printf("jsonTokener.runes: %v\n", jsonTokener.runes)
	return jsonTokener, nil
}

// back back up a rune
func (jstk *JSONTokener) back() error {
	if jstk.usePrevious || jstk.index <= 0 {
		return errors.New("不可备份2次")
	}

	jstk.decreIndex()
	jstk.usePrevious = true
	jstk.eof = false
	return nil
}

// decreIndex 当前索引减一
func (jstk *JSONTokener) decreIndex() {
	jstk.index--
	if jstk.previous == '\r' || jstk.previous == '\n' {
		jstk.line--
		jstk.runepos = jstk.runesPreviousLine
	} else if jstk.runepos > 0 {
		jstk.runepos--
	}
}

// dehexRune 获取rune的16进制值
func (jstk *JSONTokener) dehexRune(c rune) int {
	if c >= '0' && c <= '9' {
		return int(c) - '0'
	}
	if c >= 'A' && c <= 'F' {
		return int(c) - ('A' - 10)
	}
	if c >= 'a' && c <= 'f' {
		return int(c) - ('A' - 10)
	}
	return -1
}

// isEnd 输入是否结束
func (jstk *JSONTokener) isEnd() bool {
	return jstk.eof && !jstk.usePrevious
}

func (jstk *JSONTokener) hasMore() bool {
	if jstk.usePrevious {
		return true
	}

	return jstk.index < (len(jstk.runes) - 1)
}

// next 获取输入中的下一个字符
func (jstk *JSONTokener) next() rune {
	var c rune
	if jstk.usePrevious {
		jstk.usePrevious = false
		c = jstk.previous
	} else {
		if jstk.isEnd() {
			jstk.eof = true
			return '\000'
		}
		c = jstk.runes[jstk.index]
	}
	// log.Printf("next: %#U\n", c)
	jstk.increIndex(c)
	jstk.previous = c
	return jstk.previous
}

func (jstk *JSONTokener) increIndex(c rune) {
	if c != '\000' {
		jstk.index++
		if c == '\r' {
			jstk.line++
			jstk.runesPreviousLine = jstk.runepos
			jstk.runepos = 0
		} else if c == '\n' {
			if jstk.previous != '\r' {
				jstk.line++
				jstk.runesPreviousLine = jstk.runepos
			}
			jstk.runepos = 0
		} else {
			jstk.runepos++
		}
	}
}

func (jstk *JSONTokener) match(c rune) (rune, error) {
	n := jstk.next()

	if n != c {
		if n != '\000' {
			return '\000', fmt.Errorf("expected %c, given %c", c, n)
		}
		return '\000', fmt.Errorf("expect %c, given ''", c)
	}
	return n, nil
}

func (jstk *JSONTokener) getNextNRunes(n int) (string, error) {
	if n == 0 {
		return "", nil
	}

	cs := []rune{}
	pos := 0
	for pos < n {
		cs = append(cs, jstk.next())
		if jstk.isEnd() {
			return "", errors.New("获取子串出错，输入已结束")
		}
		pos++
	}

	return string(cs), nil
}

func (jstk *JSONTokener) nextNoneEmptyRune() rune {
	for {
		r := jstk.next()
		// log.Printf("nextNoneEmptyRune: %#U\n", r)
		if r == '\000' || r > ' ' {
			return r
		}
	}
}

func (jstk *JSONTokener) nextString(quote rune) (string, error) {
	var r rune
	rs := []rune{}

	for {
		r = jstk.next()

		switch r {
		case '\000', '\r', '\n':
			return "", errors.New("Unterminated string")
		case '\\':
			r = jstk.next()
			switch r {
			case 'b':
				rs = append(rs, '\b')
			case 't':
				rs = append(rs, '\t')
			case 'n':
				rs = append(rs, '\n')
			case 'f':
				rs = append(rs, '\f')
			case 'r':
				rs = append(rs, '\r')
			case 'u':
				s, err := jstk.getNextNRunes(4)
				if err != nil {
					return "", err
				}
				v, err := strconv.ParseInt(s, 10, 16)
				if err != nil {
					return "", err
				}
				rs = append(rs, rune(v))
			case '"', '\\', '\'', '/':
				rs = append(rs, r)
			default:
				return "", errors.New("illega escape")
			}
		default:
			if r == quote {
				return string(rs), nil
			}
			rs = append(rs, r)
		}
	}
}

func (jstk *JSONTokener) nextValue() (interface{}, error) {
	r := jstk.nextNoneEmptyRune()
	var s string

	switch r {
	case '"', '\'':
		ss, err := jstk.nextString(r)
		return ss, err
	case '{':
		jstk.back()
		return NewJSONObject(jstk)
	case '[':
		jstk.back()
		v, err := NewJSONArray(jstk)
		return v, err
	}

	rs := []rune{}
	for r >= ' ' && !strings.ContainsRune(",:]}/\\\"[{;=#", r) {
		rs = append(rs, r)
		r = jstk.next()
	}
	jstk.back()

	s = strings.Trim(string(rs), " ")
	// log.Printf("nextValue: %s\n", s)
	if s == "" {
		return nil, errors.New("missing value")
	}
	return stringToValue(s)
}

func (jstk *JSONTokener) skipBlankRunes() {

}
