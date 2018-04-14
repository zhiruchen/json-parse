package json_parse

import (
	"io"
	"fmt"

	"github.com/zhiruchen/json-parse/parser"
	"github.com/zhiruchen/json-parse/scanner"
	"errors"
)

var (
	scanErrorFunc = func(line int, msg string) {
		panic(fmt.Sprintf("scanner error %d: %s\n", line, msg))
	}

	parseErrorFunc = func(name, msg string) {
		panic(fmt.Sprintf("parser error %s: %s\n", name, msg))
	}
)

// GetValue get value in path
func GetValue(r io.Reader, path ...string) (interface{}, error) {
	jsonObj, err := GetJSONObject(r)
	if err != nil {
		return nil, err
	}

	if v, ok := jsonObj.(parser.JSONObject); ok {
		return v.GetValue(path...)
	}

	return nil, errors.New(fmt.Sprintf("%v is not json object", jsonObj))
}

// GetJSONObject get json object from reader
func GetJSONObject(r io.Reader) (parser.JSONer, error) {
	var bs []byte
	_, err := r.Read(bs)
	if err != nil {
		return nil, err
	}

	sc := scanner.NewScanner(string(bs), scanErrorFunc)
	tokens := sc.ScanTokens()

	p := parser.NewParser(tokens, parseErrorFunc)
	return p.Parse(), nil
}