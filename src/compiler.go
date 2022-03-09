package template

// A mode value is a set of flags (or 0). Modes control parser behavior.
type Mode uint

type state struct {
	mode Mode
}

const (
	ParseComments Mode = 1 << iota // parse comments and add them to AST
	SkipFuncCheck                  // do not check that functions are defined
)
