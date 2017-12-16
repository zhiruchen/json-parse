package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/zhiruchen/json-parse/parser"
	"github.com/zhiruchen/json-parse/scanner"
)

func readFile(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func main() {
	jsonPath := flag.String("path", "", "json文件路径")
	flag.Parse()

	s, err := readFile(*jsonPath)
	if err != nil {
		log.Fatalln(err)
	}

	scanErrfunc := func(line int, msg string) {
		log.Fatalf("scanner error %d: %s\n", line, msg)
	}

	sc := scanner.NewScanner(s, scanErrfunc)
	tokens := sc.ScanTokens()

	errFunc := func(name, msg string) {
		log.Fatalf("parser error %s: %s\n", name, msg)
	}

	pser := parser.NewParser(tokens, errFunc)
	js := pser.Parse()

	fmt.Printf(js.Represent())
}
