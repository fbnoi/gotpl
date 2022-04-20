package main

import (
	"fmt"
	"strings"

	"fbnoi.com/gotpl/template"
	"github.com/pkg/errors"
)

func main() {
	sb := &strings.Builder{}
	if err := template.Render(sb, "./cmd/test.html"); err != nil {
		panic(errors.WithStack(err))
	}
	fmt.Println(sb.String())
}
