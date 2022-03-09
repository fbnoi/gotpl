package main

import (
	"fmt"
	"gotpl/template"
)

var html = `{% if a == 1 && b == 2 %}`

func main() {
	lexer := &template.Lexer{}
	stream := lexer.Tokenize(&template.Source{Code: html})
	for !stream.IsEOF() {
		fmt.Println(stream.Next().String())
	}
	fmt.Println(stream.String())
}
