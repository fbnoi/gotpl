package main

import (
	"fmt"
	"gotpl/template"
)

// func main() {
// 	fset := token.NewFileSet()
// 	path, _ := filepath.Abs("./cmd/demo.go")
// 	f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// 	ast.Print(fset, f)
// }

func main() {
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>{{ title }}</title>
</head>
<body>
	{% set a = b.c(d.e, 1+ d.f) %}
</body>
</html>
	`

	lex := &template.Lexer{}
	stream := lex.Tokenize(&template.Source{Code: html})
	filter := &TokenFilter{tr: &Tree{}}
	tree := filter.Filter(stream)
	fmt.Println(tree)
}
