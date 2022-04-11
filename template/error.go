package template

import "fmt"

type UnexpectedEndOfFile struct {
	Path string
}
type UnexpectedToken struct {
	Path  string
	Token string
	Line  int
}
type ParseTemplateFaild struct {
	Path string
	Line int
}

func (e UnexpectedEndOfFile) Error() string {
	return fmt.Sprintf("un expected end of file at file:%s", e.Path)
}

func (e UnexpectedToken) Error() string {
	return fmt.Sprintf("un expected token \"%s\" at file:%s:%d", e.Token, e.Path, e.Line)
}

func (e ParseTemplateFaild) Error() string {
	return fmt.Sprintf("Parse Template Faild at file:%s:%d", e.Path, e.Line)
}

func NewUnexpectedEndOfFile(path string) UnexpectedEndOfFile {
	return UnexpectedEndOfFile{path}
}

func NewUnexpectedToken(path, tok string, line int) UnexpectedToken {
	return UnexpectedToken{path, tok, line}
}

func NewParseTemplateFaild(path string, line int) ParseTemplateFaild {
	return ParseTemplateFaild{path, line}
}
