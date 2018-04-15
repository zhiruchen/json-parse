package json_parse

import (
	"bytes"
	"testing"
	"reflect"

	"github.com/zhiruchen/json-parse/parser"
)

func TestGetValue(t *testing.T) {
	testCases := []struct {
		js     string
		path   []string
		result interface{}
	}{
		{
			js:     `{"A":"a", "B": {"C":1, "D":2}, "E":9}`,
			path:   []string{"A"},
			result: "a",
		},
		{
			js:     `{"A":"a", "B": {"C":1, "D":2}, "E":9}`,
			path:   []string{"B", "D"},
			result: float64(2),
		},
		{
			js:     `{"A":"a", "B": {"C":1, "D":2}, "E":9}`,
			path:   []string{"A", "B", "E"},
			result: nil,
		},
		{
			js:     `{"A":"a", "B": {"C":1, "D":2}, "E":9}`,
			path:   []string{"B", "C"},
			result: float64(1),
		},
		{
			js:     `{"A":"a", "B": {"C":1, "D":2}, "E":9}`,
			path:   []string{"A", "B", "c"},
			result: nil,
		},
	}

	for _, cc := range testCases {
		v, _ := GetValue(bytes.NewBufferString(cc.js), cc.path...)
		if v != cc.result {
			t.Errorf("expect: %v, get: %v\n", cc.result, v)
		}
	}
}

func TestGetJSONObject(t *testing.T) {
	testCases := []struct {
		js     string
		result parser.JSONer
	}{
		{
			js: `{"A":"a", "B": {"C":1, "D":2, "F":{"J":"SQ","Y":100000}}, "E":9}`,
			result: parser.JSONObject{
				"A": "a",
				"B": parser.JSONObject{
					"C": float64(1),
					"D": float64(2),
					"F": parser.JSONObject{
						"J": "SQ",
						"Y": float64(100000),
					},
				},
				"E": float64(9),
			},
		},
		{
			js: `{"A":"a", "E":9}`,
			result: parser.JSONObject{
				"A": "a",
				"E": float64(9),
			},
		},
		{
			js: `[1,2,3,"x","y", {"A":"a","B":"b","C":[11,12,13]}]`,
			result: parser.JSONArray{
				float64(1),
				float64(2),
				float64(3),
				"x",
				"y",
				parser.JSONObject{
					"A": "a",
					"B": "b",
					"C": parser.JSONArray{float64(11), float64(12), float64(13)},
				},
			},
		},
	}

	for _, cc := range testCases {
		v, _ := GetJSONObject(bytes.NewBufferString(cc.js))
		if !reflect.DeepEqual(v, cc.result) {
			t.Errorf("expect: %v, get: %v\n", cc.result, v)
		}
	}
}
