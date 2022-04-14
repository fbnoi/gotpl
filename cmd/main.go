package main

import (
	"fmt"
	"log"
	"strings"

	"fbnoi.com/gotpl/template"
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
	sb := &strings.Builder{}
	if err := template.Render(sb, "./cmd/test.html"); err != nil {
		log.Println(err)
	}
	fmt.Println(sb.String())
}
