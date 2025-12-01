package main

import (
	"flag"
	"os"
	"bufio"
	"io"
	"github.com/suhailslh/cc-json-parser/parser"
	"fmt"
)

func run() int {
	flag.Parse()

	filename := flag.Arg(0)
	file, err := os.Open(filename)
	if err != nil {
		return 1
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return 1
	}

	data := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(data)
	if err != nil && err != io.EOF {
		return 1
	}

	isValidJSON, err := parser.IsValidJSON(string(data))
	if !isValidJSON || err != nil {
		fmt.Println(err)
		return 1
	}

	fmt.Println("valid json")
	return 0
}

func main() {
	os.Exit(run())
}
