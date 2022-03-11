package main

import "fmt"

type Slice interface {
	Append(i int)
}

type ints []int

func (nl *ints) Append(i int) {
	*nl = append([]int(*nl), i)
}

type B struct {
	ints
}

func main() {
	var b Slice = &ints{}
	b.Append(1)
	fmt.Println(b)
}
