package main

import (
	"context"
)

// Foo 结构体
type Foo struct {
	i int
}

// Bar 接口
type Bar interface {
	Do(ctx context.Context) error
}

// // main方法
// func main() {
// 	a := (1 + (3 + 4*5)) * 3 * add(4, 5)
// 	if a == 5 {

// 	}
// 	c := []int{1, 2, 3}
// 	if 1 == 1 {

// 	} else if 2 == 2 {

// 	} else {

// 	}
// }

// func add(a, b int) int {
// 	return a + b
// }
