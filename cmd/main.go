package main

import (
	"fmt"
	"gotpl/template"
)

var html = `{{ 1 | echo(1) | json_encode(2) }}`

func main() {
	lexer := &template.Lexer{}
	stream := lexer.Tokenize(&template.Source{Code: html})
	for !stream.IsEOF() {
		fmt.Println(stream.Next().String())
	}
	fmt.Println(stream.String())
}
