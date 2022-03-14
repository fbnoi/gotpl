package main

import "fmt"

type A interface{}

type b struct{}

func main() {
	x := &b{}
	fmt.Printf("%p\n", x)
	y := tointerface(x)
	fmt.Printf("%p\n", y)
	fmt.Printf("%p\n", y.(*b))
}

func tointerface(x interface{}) interface{} {
	return x
}
