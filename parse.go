package json_parse

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/zhiruchen/json-parse/parser"
	"github.com/zhiruchen/json-parse/scanner"
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
	var buf = &bytes.Buffer{}
	_, err := buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	sc := scanner.NewScanner(buf.String(), scanErrorFunc)
	tokens := sc.ScanTokens()

	p := parser.NewParser(tokens, parseErrorFunc)
	return p.Parse(), nil
}
