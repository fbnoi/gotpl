package template

import "time"

// A mode value is a set of flags (or 0). Modes control parser behavior.
type Mode uint

type state struct {
	mode Mode
}

const (
	ParseComments Mode = 1 << iota // parse comments and add them to AST
	SkipFuncCheck                  // do not check that functions are defined
)

// Document is the representation of a single parsed template.
type Document struct {
	Ceil
	LastModified *time.Time
	Name         string
	Root         bool
}

func (doc *Document) Copy() *Document {
	return &Document{
		Ceil: doc.Ceil.Copy(),
	}
}
