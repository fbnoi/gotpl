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
	ch string
	at int
}

func (b *bracket) String() string {
	return fmt.Sprintf("%s at %d", b.ch, b.at)
}

type Lexer struct {
	tokens    []*Token
	source    *Source
	code      string
	cursor    int
	lineno    int
	end       int
	position  int
	positions [][]int
}

func (lex *Lexer) Tokenize(source *Source) *TokenStream {
	lex.source = source
	lex.code = reg_enter.ReplaceAllString(source.Code, "\n")
	lex.cursor = 0
	lex.lineno = 1
	lex.end = len(lex.code)
	lex.position = -1
	lex.positions = reg_token_start.FindAllStringIndex(lex.code, -1)
	if len(lex.positions) == 0 {
		lex.pushToken(TYPE_TEXT, lex.code[lex.cursor:])
		lex.cursor = lex.end
	}

	for lex.position < len(lex.positions)-1 {
		lex.position++
		lex.lexFromTo()
	}
	if lex.cursor < lex.end {
		lex.pushToken(TYPE_TEXT, lex.code[lex.cursor:lex.end])
	}
	lex.pushToken(TYPE_EOF, "")
	return &TokenStream{tokens: lex.tokens, source: lex.source}
}

func (lex *Lexer) lexFromTo() {
	position := lex.positions[lex.position]
	if position[0] < lex.cursor {
		return
	} else if position[0] > lex.cursor {
		lex.pushToken(TYPE_TEXT, lex.code[lex.cursor:position[0]])
	}
	lex.moveCursor(position[1])
	switch position[1] - position[0] {
	case 3:
		var (
			reg *regexp.Regexp
			tag string
		)
		switch lex.code[position[0]+1 : position[1]] {
		case TAG_COMMENT[0]:
			reg = reg_comment
			tag = TAG_COMMENT[0]
		case TAG_BLOCK[0]:
			reg = reg_block
			tag = TAG_BLOCK[0]
		case TAG_VARIABLE[0]:
			reg = reg_variable
			tag = TAG_VARIABLE[0]
		}
		if !reg_variable.MatchString(lex.code[lex.cursor:]) {
			panic(NewErrorf(lex.lineno, lex.source, "Unclosed %s", tag))
		}
		subp := lex.findStringIndex(reg, lex.code, lex.cursor)
		lex.pushToken(TYPE_TEXT, lex.code[position[0]+1:subp[1]])
		lex.moveCursor(subp[1])
	case 2:
		switch lex.code[position[0]:position[1]] {
		case TAG_COMMENT[0]:
			lex.lexComment()
		case TAG_BLOCK[0]:
			if subp := lex.findStringIndex(reg_block_raw, lex.code, lex.cursor); len(subp) > 0 && subp[0] == lex.cursor {
				position := lex.findStringIndex(reg_block_raw, lex.code, lex.cursor)
				lex.moveCursor(position[1])
				subp := lex.findStringIndex(reg_raw_data, lex.code, lex.cursor)
				if len(subp) == 0 {
					panic(NewError(lex.lineno, lex.source, "Unclosed {% verbatim %}"))
				}
				lex.pushToken(TYPE_STRING, lex.code[lex.cursor:subp[0]])
				lex.moveCursor(subp[1])
			} else {
				lex.pushToken(TYPE_BLOCK_START, "")
				lex.LexRegData(reg_block, TAG_BLOCK[0])
				lex.pushToken(TYPE_BLOCK_END, "")
			}
		case TAG_VARIABLE[0]:
			lex.pushToken(TYPE_VAR_START, "")
			lex.LexRegData(reg_variable, TAG_VARIABLE[0])
			lex.pushToken(TYPE_VAR_END, "")
		}
	}
}

func (lex *Lexer) lexComment() {
	if !reg_comment.MatchString(lex.code[lex.cursor:]) {
		panic(NewErrorf(lex.lineno, lex.source, "Unclosed %s", TAG_COMMENT[0]))
	}
	position := lex.findStringIndex(reg_comment, lex.code, lex.cursor)
	lex.moveCursor(position[1])
}

func (lex *Lexer) LexRegData(reg *regexp.Regexp, tag string) {
	if !reg.MatchString(lex.code[lex.cursor:]) {
		panic(NewErrorf(lex.lineno, lex.source, "Unclosed %s", tag))
	}
	lex.lexExpression(reg)
}

func (lex *Lexer) lexExpression(reg *regexp.Regexp) {
	position := lex.findStringIndex(reg, lex.code, lex.cursor)
	var brackets []*bracket

	for lex.cursor < position[0] {
		// whitespace
		if reg_whitespace.MatchString(lex.code[lex.cursor:]) {
			subp := lex.findStringIndex(reg_whitespace, lex.code[:position[0]], lex.cursor)
			lex.moveCursor(subp[1])
		}
		// operator
		if subp := lex.findStringIndex(reg_operator, lex.code[:position[0]], lex.cursor); len(subp) > 0 && subp[0] == lex.cursor {
			lex.pushToken(TYPE_OPERATOR, lex.code[lex.cursor:subp[1]])
			lex.moveCursor(subp[1])

			// name
		} else if subp := lex.findStringIndex(reg_name, lex.code[:position[0]], lex.cursor); len(subp) > 0 && subp[0] == lex.cursor {
			lex.pushToken(TYPE_NAME, lex.code[lex.cursor:subp[1]])
			lex.moveCursor(subp[1])

			// number
		} else if subp := lex.findStringIndex(reg_number, lex.code[:position[0]], lex.cursor); len(subp) > 0 && subp[0] == lex.cursor {
			lex.pushToken(TYPE_NUMBER, lex.code[lex.cursor:subp[1]])
			lex.moveCursor(subp[1])

			// string
		} else if subp := lex.findStringIndex(reg_string, lex.code[:position[0]], lex.cursor); len(subp) > 0 && subp[0] == lex.cursor {
			lex.pushToken(TYPE_STRING, lex.code[lex.cursor:subp[1]])
			lex.moveCursor(subp[1])

			// punctuation
		} else if subp := lex.findStringIndex(reg_punctuation, lex.code[:position[0]], lex.cursor); len(subp) > 0 && subp[0] == lex.cursor {
			var subp []int

			// bracket open
			if subp = lex.findStringIndex(reg_bracket_open, lex.code[:position[0]], lex.cursor); len(subp) > 0 && subp[0] == lex.cursor {
				brackets = append(brackets, &bracket{ch: lex.code[subp[0]:subp[1]], at: subp[0]})

				// bracket close
			} else if subp = lex.findStringIndex(reg_bracket_close, lex.code[:position[0]], lex.cursor); len(subp) > 0 && subp[0] == lex.cursor {
				if len(brackets) == 0 {
					panic(NewErrorf(lex.lineno, lex.source, "Unexpected token %s", lex.code[subp[0]:subp[1]]))
				}
				b := brackets[len(brackets)-1]
				switch b.ch {
				case "{":
					if lex.code[subp[0]:subp[1]] != "}" {
						panic(NewErrorf(lex.lineno, lex.source, "Unexpected token %s", lex.code[subp[0]:subp[1]]))
					}
				case "(":
					if lex.code[subp[0]:subp[1]] != ")" {
						panic(NewErrorf(lex.lineno, lex.source, "Unexpected token %s", lex.code[subp[0]:subp[1]]))
					}
				case "[":
					if lex.code[subp[0]:subp[1]] != "]" {
						panic(NewErrorf(lex.lineno, lex.source, "Unexpected token %s", lex.code[subp[0]:subp[1]]))
					}
				default:
					panic(NewErrorf(lex.lineno, lex.source, "Unexpected token %s", lex.code[subp[0]:subp[1]]))
				}
				brackets = brackets[:len(brackets)-1]
			} else {
				subp = lex.findStringIndex(reg_punctuation, lex.code[:position[0]], lex.cursor)
			}
			lex.pushToken(TYPE_PUNCTUATION, lex.code[lex.cursor:subp[1]])
			if len(subp) > 0 {
				lex.moveCursor(subp[1])
			}
		} else {
			// unkown token
			panic(NewErrorf(lex.lineno, lex.source, "Unexpected token %s", lex.code[lex.cursor:lex.cursor+1]))
		}
	}

	if len(brackets) > 0 {
		panic(NewErrorf(lex.lineno, lex.source, "Unclosed %s", brackets[len(brackets)-1]))
	}
	lex.moveCursor(position[1])
}

func (lex *Lexer) findStringIndex(reg *regexp.Regexp, str string, offset int) []int {
	position := reg.FindStringIndex(str[offset:])
	if len(position) == 0 {
		return []int{}
	}
	return []int{position[0] + offset, position[1] + offset}
}

func (lex *Lexer) pushToken(typ int, value string) {
	if typ == TYPE_TEXT && value == "" {
		return
	}
	lex.tokens = append(lex.tokens, &Token{typ: typ, value: value, line: lex.lineno, at: lex.cursor})
}

func (lex *Lexer) moveCursor(n int) {
	lex.lineno += len(reg_enter.FindAllString(lex.code[lex.cursor:n], -1))
	lex.cursor = n
}
