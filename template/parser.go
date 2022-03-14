package template

import "strings"

type Parser struct {
	nodeStack []BranchAble
	stream    *TokenStream
	upper     BranchAble
	cursor    Node
	doc       *Document
}

func (p *Parser) Parse(stream *TokenStream) *Document {
	p.stream = stream
	p.doc = &Document{}
	p.upper = p.doc
	for !stream.IsEOF() {
		token := stream.Next()
		switch token.Type() {
		case TYPE_TEXT:
			p.parseText(token)
		case TYPE_BLOCK_START, TYPE_BLOCK_END:
			p.parseBlock(token)
		case TYPE_VAR_START, TYPE_VAR_END:
		case TYPE_NAME, TYPE_NUMBER, TYPE_STRING, TYPE_OPERATOR, TYPE_PUNCTUATION:
			p.parsePipeNode(token)
		default:
		}
	}
	return p.doc
}

func (p *Parser) parseText(token *Token) {
	p.upper.Append(newTextNode(token.Value(), token.At))
}
func (p *Parser) parseBlock(token *Token) {
	if token.Type() == TYPE_BLOCK_START {
		token = p.stream.Next()
		switch token.Value() {
		case "if":
			p.parseIfNode(token)
		case "elseif":
			p.parseElseIfNode(token)
		case "else":
			p.parseElseNode(token)
		case "endif":
			p.parseElseNode(token)
		case "set":
		case "for":
		case "endfor":
		case "block":
		case "endblock":
		case "extend":
		case "include":
		}
	}
}

func (p *Parser) parseIfNode(token *Token) {
	node := newIfNode(token.At, p.doc)
	pos := token.At
	sb := &strings.Builder{}
	for token.Type() != TYPE_BLOCK_END {
		sb.WriteString(token.Value())
		token = p.stream.Next()
	}
	node.PipeNode = newPipeNode(sb.String(), pos)
	p.cursor = node
	p.upper.Append(node)
	p.nodeStack = append(p.nodeStack, p.upper)
	p.upper = node
}

func (p *Parser) parseElseIfNode(token *Token) {
	node := newElseIfNode(token.At)
	if p.cursor != nil {
		if p.cursor.Type() != NodeElseif || p.cursor.Type() != NodeIf {
			panic("")
		}
	}
	pos := token.At
	sb := &strings.Builder{}
	for token.Type() != TYPE_BLOCK_END {
		sb.WriteString(token.Value())
		token = p.stream.Next()
	}
	node.PipeNode = newPipeNode(sb.String(), pos)
	p.cursor = node
}

func (p *Parser) parseElseNode(token *Token) {
	node := newIfNode(token.At, p.doc)
	if p.cursor != nil {
		if p.cursor.Type() != NodeElseif || p.cursor.Type() != NodeIf {
			panic("")
		}
	}
	p.cursor = node
}

func (p *Parser) parsePipeNode(token *Token) {
	pos := token.At
	sb := &strings.Builder{}
	for token.Type() != TYPE_BLOCK_END {
		sb.WriteString(token.Value())
		token = p.stream.Next()
	}
	node := newPipeNode(sb.String(), pos)
	switch p.cursor.Type() {
	case NodeIf:
		p.cursor.(*IfNode).PipeNode = node
	case NodeElseif:
		p.cursor.(*ElseIfNode).PipeNode = node
	case NodeSet:
		p.cursor.(*SetNode).PipeNode = node
	case NodeExtend:
		// p.cursor.().PipeNode = node
	case NodeInclude:
		// p.cursor.(*SetNode).PipeNode = node
	case NodeRange:
		p.cursor.(*RangeNode).PipeNode = node
	}
}

func (p *Parser) pushStack() {
	if b, ok := p.cursor.(BranchAble); ok {
		p.upper.Append(p.cursor)
		p.nodeStack = append(p.nodeStack, p.upper)
		p.upper = b
	}
}

func (p *Parser) popStack() {

}
