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
	// {% endverbatim %}
	reg_raw_data = regexp.MustCompile(fmt.Sprintf(`%s\s*endverbatim\s*%s`, TAG_BLOCK[0], TAG_BLOCK[1]))
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

type bracket struct {
	ch   string
	line int
}

func (b *bracket) String() string {
	return fmt.Sprintf("%s at line %d", b.ch, b.line)
}

func Lexer() *lexer {
	return &lexer{}
}

type lexer struct {
	tpl    *Template
	tokens []*Token
	code   string
	cursor int
	line   int
	end    int
	posIdx int
	poss   [][]int
}

func (lex *lexer) Tokenize(str string) (*TokenStream, error) {
	lex.code = reg_enter.ReplaceAllString(str, "\n")
	lex.cursor = 0
	lex.line = 1
	lex.end = len(lex.code)
	lex.posIdx = -1
	lex.poss = reg_token_start.FindAllStringIndex(lex.code, -1)
	if len(lex.poss) == 0 {
		lex.pushToken(TYPE_TEXT, lex.code[lex.cursor:])
		lex.cursor = lex.end
	}

	for lex.posIdx < len(lex.poss)-1 {
		lex.posIdx++
		if err := lex.lexNextPart(); err != nil {
			return nil, err
		}
	}
	if lex.cursor < lex.end {
		lex.pushToken(TYPE_TEXT, lex.code[lex.cursor:lex.end])
	}
	lex.pushToken(TYPE_EOF, "")
	return &TokenStream{tokens: lex.tokens}, nil
}

func (lex *lexer) lexNextPart() error {
	pos := lex.poss[lex.posIdx]
	if pos[0] < lex.cursor {
		return nil
	} else if pos[0] > lex.cursor {
		lex.pushToken(TYPE_TEXT, lex.code[lex.cursor:pos[0]])
	}
	lex.moveCursor(pos[1])
	switch pos[1] - pos[0] {
	case 3:
		var reg *regexp.Regexp
		switch lex.code[pos[0]+1 : pos[1]] {
		case TAG_COMMENT[0]:
			reg = reg_comment
		case TAG_BLOCK[0]:
			reg = reg_block
		case TAG_VARIABLE[0]:
			reg = reg_variable
		}
		if subp, ok := startWith(reg, lex.code, lex.cursor); ok {
			lex.pushToken(TYPE_TEXT, lex.code[pos[0]+1:subp[1]])
			lex.moveCursor(subp[1])
			return nil
		}
		return NewUnexpectedToken(lex.tpl.Name, lex.code[pos[0]:pos[1]], lex.line)
	case 2:
		switch lex.code[pos[0]:pos[1]] {
		case TAG_COMMENT[0]:
			return lex.lexComment()
		case TAG_BLOCK[0]:
			if subp, ok := startWith(reg_block_raw, lex.code, lex.cursor); ok {
				lex.moveCursor(subp[1])
				if subp = findStringIndex(reg_raw_data, lex.code, lex.cursor); len(subp) > 0 {
					lex.pushToken(TYPE_STRING, lex.code[lex.cursor:subp[0]])
					lex.moveCursor(subp[1])
					return nil
				}
				return NewUnexpectedToken(lex.tpl.Name, lex.code[pos[0]:pos[1]], lex.line)
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

func (lex *lexer) lexComment() error {
	if p := findStringIndex(reg_comment, lex.code, lex.cursor); len(p) > 0 {
		lex.moveCursor(p[1])
		return nil
	}
	return NewParseTemplateFaild(lex.tpl.Name, lex.line)
}

func (lex *lexer) LexRegData(reg *regexp.Regexp) error {
	if !reg.MatchString(lex.code[lex.cursor:]) {
		return NewParseTemplateFaild(lex.tpl.Name, lex.line)
	}
	lex.lexExpression(reg)
	return nil
}

func (lex *lexer) lexExpression(reg *regexp.Regexp) error {
	pos := findStringIndex(reg, lex.code, lex.cursor)
	var brackets []*bracket

	for lex.cursor < pos[0] {
		// whitespace
		if subp, ok := startWith(reg_whitespace, lex.code[:pos[0]], lex.cursor); ok {
			lex.moveCursor(subp[1])
		}
		// operator
		if subp, ok := startWith(reg_operator, lex.code[:pos[0]], lex.cursor); ok {
			lex.pushToken(TYPE_OPERATOR, lex.code[lex.cursor:subp[1]])
			lex.moveCursor(subp[1])

			// name
		} else if subp, ok := startWith(reg_name, lex.code[:pos[0]], lex.cursor); ok {
			lex.pushToken(TYPE_NAME, lex.code[lex.cursor:subp[1]])
			lex.moveCursor(subp[1])

			// number
		} else if subp, ok := startWith(reg_number, lex.code[:pos[0]], lex.cursor); ok {
			lex.pushToken(TYPE_NUMBER, lex.code[lex.cursor:subp[1]])
			lex.moveCursor(subp[1])

			// string
		} else if subp, ok := startWith(reg_string, lex.code[:pos[0]], lex.cursor); ok {
			lex.pushToken(TYPE_STRING, lex.code[lex.cursor:subp[1]])
			lex.moveCursor(subp[1])

			// punctuation
		} else if _, ok := startWith(reg_punctuation, lex.code[:pos[0]], lex.cursor); ok {
			var (
				subp []int
				ok   bool
			)
			// bracket open
			if subp, ok = startWith(reg_bracket_open, lex.code[:pos[0]], lex.cursor); ok {
				brackets = append(brackets, &bracket{ch: lex.code[subp[0]:subp[1]], line: lex.line})

				// bracket close
			} else if subp, ok = startWith(reg_bracket_close, lex.code[:pos[0]], lex.cursor); ok {
				if len(brackets) == 0 {
					return NewParseTemplateFaild(lex.tpl.Name, lex.line)
				}
				b := brackets[len(brackets)-1]
				switch {
				case b.ch == "{" && lex.code[subp[0]:subp[1]] != "}":
					return NewParseTemplateFaild(lex.tpl.Name, lex.line)
				case b.ch == "(" && lex.code[subp[0]:subp[1]] != "}":
					return NewParseTemplateFaild(lex.tpl.Name, lex.line)
				case b.ch == "[" && lex.code[subp[0]:subp[1]] != "}":
					return NewParseTemplateFaild(lex.tpl.Name, lex.line)
				}
				brackets = brackets[:len(brackets)-1]
			} else {
				subp = findStringIndex(reg_punctuation, lex.code[:pos[0]], lex.cursor)
			}
			lex.pushToken(TYPE_PUNCTUATION, lex.code[lex.cursor:subp[1]])
			if len(subp) > 0 {
				lex.moveCursor(subp[1])
			}
		} else {
			// unkown token
			return NewUnexpectedToken(lex.tpl.Name, lex.code[lex.cursor:pos[0]], brackets[0].line)
		}
	}

	if len(brackets) > 0 {
		return NewUnexpectedToken(lex.tpl.Name, brackets[0].ch, brackets[0].line)
	}
	lex.moveCursor(pos[1])
	return nil
}

func (lex *lexer) pushToken(typ int, value string) {
	if typ == TYPE_TEXT && value == "" {
		return
	}
	lex.tokens = append(lex.tokens, &Token{typ: typ, value: value, line: lex.line, at: lex.cursor})
}

func (lex *lexer) moveCursor(n int) {
	lex.line += len(reg_enter.FindAllString(lex.code[lex.cursor:n], -1))
	lex.cursor = n
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
