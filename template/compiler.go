package template

import "time"

// Document is the representation of a single parsed template.
type Document struct {
	LastModified *time.Time
	Name         string
	Root         bool
	Branchs      []Node
}

func (doc *Document) Append(n Node) {
	doc.Branchs = append(doc.Branchs, n)
}

func (doc *Document) Walk(fn func(n Node)) {
	for _, v := range doc.Branchs {
		fn(v)
	}
}
