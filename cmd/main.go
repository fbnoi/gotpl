package main

import (
	"fmt"
	"strings"

	"fbnoi.com/gotpl/template"
)

func main() {
	sb := &strings.Builder{}
	if err := template.Render(sb, "./test.html"); err != nil {
		panic(err)
	}
	fmt.Println(sb.String())
}
