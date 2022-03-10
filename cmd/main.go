package main

import "fmt"

type A struct {
	Name string
}

func (a *A) Copy() *A {
	return &A{Name: a.Name}
}

type B struct {
	A
}

func (b *B) Copy() *B {
	return &B{A: *b.A.Copy()}
}

type C struct {
	*A
}

func (c *C) Copy() *C {
	return &C{A: c.A.Copy()}
}

func main() {
	a := &A{"111"}
	b := &B{A: *a}
	c := &C{A: a}
	fmt.Println(b.A.Name)
	fmt.Println(c.A.Name)
	a.Name = "222"
	fmt.Println(b.A.Name)
	fmt.Println(c.A.Name)
	d := b.Copy()
	e := c.Copy()
	b.A.Name = "333"
	c.A.Name = "444"
	fmt.Println(d.A.Name)
	fmt.Println(e.A.Name)
}
