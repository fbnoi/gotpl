package template

import "fmt"

func NewErrorf(line int, source *Source, format string, v ...interface{}) Error {
	var msg string
	if len(v) == 0 {
		msg = format
	} else {
		msg = fmt.Sprintf(format, v...)
	}
	return Error{
		message: msg,
		line:    line,
		source:  source,
	}
}

func NewError(line int, source *Source, msg string) Error {
	return Error{
		message: msg,
		line:    line,
		source:  source,
	}
}

type Error struct {
	message string
	line    int
	source  *Source
}

func (e Error) Error() string {
	return e.message
}

func (e Error) Line() int {
	return e.line
}

func (e Error) Source() *Source {
	return e.source
}
