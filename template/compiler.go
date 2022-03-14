package template

import (
	"strings"
	"time"
)

// Document is the representation of a single parsed template.
type Document struct {
	NodeType
	Pos
	LastModified *time.Time
	Name         string
	Root         bool
	Branchs      []Node
}

func (doc *Document) writeTo(sb *strings.Builder) {
	sb.WriteString(doc.String())
}

func (doc *Document) Copy() Node {
	return doc
}

func (doc *Document) String() string {
	return ""
}

func (doc *Document) Append(n Node) {
	doc.Branchs = append(doc.Branchs, n)
}

func (doc *Document) Walk(fn func(n Node)) {
	for _, v := range doc.Branchs {
		fn(v)
	}
}
