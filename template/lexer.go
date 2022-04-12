package template

import (
	"fmt"
	"regexp"
)

var (
	TAG_COMMENT  = [...]string{`{#`, `#}`}
	TAG_BLOCK    = [...]string{`{%`, `%}`}
	TAG_VARIABLE = [...]string{`{{`, `}}`}
)

var (
	operator = [...]string{
		"+", "-", "*", "%", "/",
		"+=", "-=", "++", "--",
		">", "<", ">=", "<=", "&&", "^", "||",
		">>", "<<", "&", "|", ">>=", "<<=", "&=", "|=",
		"or", "and", "is",
	}
)

var (
	// }}
	reg_variable = regexp.MustCompile(fmt.Sprintf(`\s*%s`, TAG_VARIABLE[1]))
	// %}
	reg_block = regexp.MustCompile(fmt.Sprintf(`\s*%s`, TAG_BLOCK[1]))
	// {% Endverbatim %}
	reg_raw_data = regexp.MustCompile(fmt.Sprintf(`%s\s*Endverbatim\s*%s`, TAG_BLOCK[0], TAG_BLOCK[1]))
	// #}
	reg_comment = regexp.MustCompile(fmt.Sprintf(`\s*%s`, TAG_COMMENT[1]))
	// verbatim %}
	reg_block_raw = regexp.MustCompile(fmt.Sprintf(`\s*verbatim\s*%s`, TAG_BLOCK[1]))
	// {{ or {% or {#
	reg_token_start = regexp.MustCompile(fmt.Sprintf(`(@?%s|@?%s|@?%s)`, TAG_VARIABLE[0], TAG_BLOCK[0], TAG_COMMENT[0]))
	// \r\n \n
	reg_enter = regexp.MustCompile(`(\r\n|\n)`)
	// whitespace
	reg_whitespace = regexp.MustCompile(`^\s+`)
	// +-*/%&^|><=
	reg_operator = regexp.MustCompile(`[\+\-\&*\/%\^><=:]{1,3}|(and)|(or)`)
	// name
	reg_name = regexp.MustCompile(`[a-zA-Z_\x7f-\xff][a-zA-Z0-9_\x7f-\xff]*(\.[a-zA-Z_\x7f-\xff][a-zA-Z0-9_\x7f-\xff]*)*`)
	// number
	reg_number = regexp.MustCompile(`[0-9]+(?:\.[0-9]+)?([Ee][\+\-][0-9]+)?`)
	// punctuation
	reg_punctuation   = regexp.MustCompile(`[\(\)\[\]\{\}\?\:;,\|]`)
	reg_bracket_open  = regexp.MustCompile(`[\{\[\(]`)
	reg_bracket_close = regexp.MustCompile(`[\}\]\)]`)
	// string
	reg_string = regexp.MustCompile(`"([^"\\\\]*(?:\\\\.[^"\\\\]*)*)"|'([^\'\\\\]*(?:\\\\.[^\'\\\\]*)*)'`)
)

type Bracket struct {
	ch   string
	Line int
}

func (b *Bracket) String() string {
	return fmt.Sprintf("%s at line %d", b.ch, b.Line)
}

func NewLexer() *Lexer {
	return &Lexer{}
}

type Lexer struct {
	Source *Source
	Tokens []*Token
	Code   string
	Cursor int
	Line   int
	End    int
	PosIdx int
	Poss   [][]int
}

func (lex *Lexer) Tokenize(src *Source) (*TokenStream, error) {
	lex.Source = src
	lex.Code = reg_enter.ReplaceAllString(src.Code, "\n")
	lex.Cursor = 0
	lex.Line = 1
	lex.End = len(lex.Code)
	lex.PosIdx = -1
	lex.Poss = reg_token_start.FindAllStringIndex(lex.Code, -1)
	if len(lex.Poss) == 0 {
		lex.pushToken(TYPE_TEXT, lex.Code[lex.Cursor:])
		lex.Cursor = lex.End
	}

	for lex.PosIdx < len(lex.Poss)-1 {
		lex.PosIdx++
		if err := lex.lexNextPart(); err != nil {
			return nil, err
		}
	}
	if lex.Cursor < lex.End {
		lex.pushToken(TYPE_TEXT, lex.Code[lex.Cursor:lex.End])
	}
	lex.pushToken(TYPE_EOF, "")
	return &TokenStream{tokens: lex.Tokens}, nil
}

func (lex *Lexer) lexNextPart() error {
	pos := lex.Poss[lex.PosIdx]
	if pos[0] < lex.Cursor {
		return nil
	} else if pos[0] > lex.Cursor {
		lex.pushToken(TYPE_TEXT, lex.Code[lex.Cursor:pos[0]])
	}
	lex.moveCursor(pos[1])
	switch pos[1] - pos[0] {
	case 3:
		var reg *regexp.Regexp
		switch lex.Code[pos[0]+1 : pos[1]] {
		case TAG_COMMENT[0]:
			reg = reg_comment
		case TAG_BLOCK[0]:
			reg = reg_block
		case TAG_VARIABLE[0]:
			reg = reg_variable
		}
		if subp, ok := startWith(reg, lex.Code, lex.Cursor); ok {
			lex.pushToken(TYPE_TEXT, lex.Code[pos[0]+1:subp[1]])
			lex.moveCursor(subp[1])
			return nil
		}
		return NewUnexpectedToken(lex.Source, lex.Line, lex.Code[pos[0]:pos[1]])
	case 2:
		switch lex.Code[pos[0]:pos[1]] {
		case TAG_COMMENT[0]:
			return lex.lexComment()
		case TAG_BLOCK[0]:
			if subp, ok := startWith(reg_block_raw, lex.Code, lex.Cursor); ok {
				lex.moveCursor(subp[1])
				if subp = findStringIndex(reg_raw_data, lex.Code, lex.Cursor); len(subp) > 0 {
					lex.pushToken(TYPE_STRING, lex.Code[lex.Cursor:subp[0]])
					lex.moveCursor(subp[1])
					return nil
				}
				return NewUnexpectedToken(lex.Source, lex.Line, lex.Code[pos[0]:pos[1]])
			} else {
				lex.pushToken(TYPE_BLOCK_START, "")
				if err := lex.LexRegData(reg_block); err != nil {
					return err
				}
				lex.pushToken(TYPE_BLOCK_END, "")
				return nil
			}
		case TAG_VARIABLE[0]:
			lex.pushToken(TYPE_VAR_START, "")
			if err := lex.LexRegData(reg_variable); err != nil {
				return err
			}
			lex.pushToken(TYPE_VAR_END, "")
			return nil
		}
	}
	return nil
}

func (lex *Lexer) lexComment() error {
	if p := findStringIndex(reg_comment, lex.Code, lex.Cursor); len(p) > 0 {
		lex.moveCursor(p[1])
		return nil
	}
	return NewParseTemplateFaild(lex.Source, lex.Line)
}

func (lex *Lexer) LexRegData(reg *regexp.Regexp) error {
	if !reg.MatchString(lex.Code[lex.Cursor:]) {
		return NewParseTemplateFaild(lex.Source, lex.Line)
	}
	lex.lexExpression(reg)
	return nil
}

func (lex *Lexer) lexExpression(reg *regexp.Regexp) error {
	pos := findStringIndex(reg, lex.Code, lex.Cursor)
	var brackets []*Bracket

	for lex.Cursor < pos[0] {
		// whitespace
		if subp, ok := startWith(reg_whitespace, lex.Code[:pos[0]], lex.Cursor); ok {
			lex.moveCursor(subp[1])
		}
		// operator
		if subp, ok := startWith(reg_operator, lex.Code[:pos[0]], lex.Cursor); ok {
			lex.pushToken(TYPE_OPERATOR, lex.Code[lex.Cursor:subp[1]])
			lex.moveCursor(subp[1])

			// name
		} else if subp, ok := startWith(reg_name, lex.Code[:pos[0]], lex.Cursor); ok {
			lex.pushToken(TYPE_NAME, lex.Code[lex.Cursor:subp[1]])
			lex.moveCursor(subp[1])

			// number
		} else if subp, ok := startWith(reg_number, lex.Code[:pos[0]], lex.Cursor); ok {
			lex.pushToken(TYPE_NUMBER, lex.Code[lex.Cursor:subp[1]])
			lex.moveCursor(subp[1])

			// string
		} else if subp, ok := startWith(reg_string, lex.Code[:pos[0]], lex.Cursor); ok {
			lex.pushToken(TYPE_STRING, lex.Code[lex.Cursor:subp[1]])
			lex.moveCursor(subp[1])

			// punctuation
		} else if _, ok := startWith(reg_punctuation, lex.Code[:pos[0]], lex.Cursor); ok {
			var (
				subp []int
				ok   bool
			)
			// bracket open
			if subp, ok = startWith(reg_bracket_open, lex.Code[:pos[0]], lex.Cursor); ok {
				brackets = append(brackets, &Bracket{ch: lex.Code[subp[0]:subp[1]], Line: lex.Line})

				// bracket close
			} else if subp, ok = startWith(reg_bracket_close, lex.Code[:pos[0]], lex.Cursor); ok {
				if len(brackets) == 0 {
					return NewParseTemplateFaild(lex.Source, lex.Line)
				}
				b := brackets[len(brackets)-1]
				switch {
				case b.ch == "{" && lex.Code[subp[0]:subp[1]] != "}":
					return NewParseTemplateFaild(lex.Source, lex.Line)
				case b.ch == "(" && lex.Code[subp[0]:subp[1]] != "}":
					return NewParseTemplateFaild(lex.Source, lex.Line)
				case b.ch == "[" && lex.Code[subp[0]:subp[1]] != "}":
					return NewParseTemplateFaild(lex.Source, lex.Line)
				}
				brackets = brackets[:len(brackets)-1]
			} else {
				subp = findStringIndex(reg_punctuation, lex.Code[:pos[0]], lex.Cursor)
			}
			lex.pushToken(TYPE_PUNCTUATION, lex.Code[lex.Cursor:subp[1]])
			if len(subp) > 0 {
				lex.moveCursor(subp[1])
			}
		} else {
			// unkown token
			return NewUnexpectedToken(lex.Source, brackets[0].Line, lex.Code[lex.Cursor:pos[0]])
		}
	}

	if len(brackets) > 0 {
		return NewUnexpectedToken(lex.Source, brackets[0].Line, brackets[0].ch)
	}
	lex.moveCursor(pos[1])
	return nil
}

func (lex *Lexer) pushToken(typ int, value string) {
	if typ == TYPE_TEXT && value == "" {
		return
	}
	lex.Tokens = append(lex.Tokens, &Token{typ: typ, value: value, line: lex.Line, at: lex.Cursor})
}

func (lex *Lexer) moveCursor(n int) {
	lex.Line += len(reg_enter.FindAllString(lex.Code[lex.Cursor:n], -1))
	lex.Cursor = n
}

func startWith(reg *regexp.Regexp, str string, offset int) ([]int, bool) {
	pos := findStringIndex(reg, str, offset)
	if len(pos) == 0 {
		return []int{}, false
	}
	if pos[0] == offset {
		return pos, true
	}
	return pos, false
}

func findStringIndex(reg *regexp.Regexp, str string, offset int) []int {
	pos := reg.FindStringIndex(str[offset:])
	if len(pos) == 0 {
		return []int{}
	}
	return []int{pos[0] + offset, pos[1] + offset}
}
