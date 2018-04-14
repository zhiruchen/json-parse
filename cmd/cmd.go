package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/zhiruchen/json-parse"
)

func main() {
	jsonPath := flag.String("path", "", "json文件路径")
	flag.Parse()

	f, err := os.Open(*jsonPath)
	if err != nil {
		log.Fatalf("open file error: %v\n", err)
	}
	defer f.Close()

	js, err := json_parse.GetJSONObject(f)
	if err != nil {
		log.Fatalf("parse json error: %v\n", err)
	}

	fmt.Printf(js.Represent())
}
