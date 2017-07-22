package main

import (
	"flag"
	"io/ioutil"
	"log"

	"fmt"

	"github.com/zhiruchen/json-parse/parse"
)

func readFile(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func main() {
	jsonPath := flag.String("jsonPath", "", "json文件路径")
	flag.Parse()

	s, err := readFile(*jsonPath)
	if err != nil {
		log.Fatalln(err)
	}
	tokener, err1 := parse.NewJSONTokener(s)
	if err1 != nil {
		log.Fatalln(err1)
	}
	obj, err2 := parse.NewJSONObject(tokener)
	if err2 != nil {
		log.Fatalln(err2)
	}
	log.Printf("%v\n", obj)

	v, err3 := obj.GetValue("a")
	fmt.Printf("%v, %v\n", v, err3)

	v, err3 = obj.GetValue("e.h.q")
	fmt.Printf("%v, %v\n", v, err3)
}
