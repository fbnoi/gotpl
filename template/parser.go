package template

import "strings"

type Parser struct {
	nodeStack []Node
	stream    *TokenStream
	floor     Node
	doc       *Document
}

func (p *Parser) Parse(stream *TokenStream) *Document {
	p.stream = stream
	p.doc = &Document{}
	for !stream.IsEOF() {
		token := stream.Next()
		switch token.Type() {
		case TYPE_TEXT:
			p.parseText(token)
		case TYPE_BLOCK_START, TYPE_BLOCK_END:
			p.parseBlock(token)
		case TYPE_VAR_START:
		case TYPE_VAR_END:
		case TYPE_NAME, TYPE_NUMBER, TYPE_STRING, TYPE_OPERATOR, TYPE_PUNCTUATION:
			p.parsePipeNode(token)
		default:
		}
	}
	return p.doc
}

func (p *Parser) parseText(token *Token) {
	// textNode := newTextNode(token.Value(), token.At, p.doc)
	// p.ceil.Hang(textNode)
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
	if p.floor != nil {
		p.nodeStack = append(p.nodeStack, p.floor)
	}

	pos := token.At
	sb := &strings.Builder{}
	for token.Type() != TYPE_BLOCK_END {
		sb.WriteString(token.Value())
		token = p.stream.Next()
	}
	node.PipeNode = newPipeNode(sb.String(), pos)
	p.floor = node
}

func (p *Parser) parseElseIfNode(token *Token) {
	node := newElseIfNode(token.At, p.doc)
	if p.floor != nil {
		if p.floor.Type() != NodeElseif || p.floor.Type() != NodeIf {
			panic("")
		}
		p.nodeStack = append(p.nodeStack, p.floor)
	}
	pos := token.At
	sb := &strings.Builder{}
	for token.Type() != TYPE_BLOCK_END {
		sb.WriteString(token.Value())
		token = p.stream.Next()
	}
	node.PipeNode = newPipeNode(sb.String(), pos)
	p.floor = node
}

func (p *Parser) parseElseNode(token *Token) {
	node := newIfNode(token.At, p.doc)
	if p.floor != nil {
		if p.floor.Type() != NodeElseif || p.floor.Type() != NodeIf {
			panic("")
		}
		p.nodeStack = append(p.nodeStack, p.floor)
	}
	p.floor = node
}

func (p *Parser) parsePipeNode(token *Token) {
	pos := token.At
	sb := &strings.Builder{}
	for token.Type() != TYPE_BLOCK_END {
		sb.WriteString(token.Value())
		token = p.stream.Next()
	}
	node := newPipeNode(sb.String(), pos)
	switch p.floor.Type() {
	case NodeIf:
		p.floor.(*IfNode).PipeNode = node
	case NodeElseif:
		p.floor.(*ElseIfNode).PipeNode = node
	case NodeSet:
		p.floor.(*SetNode).PipeNode = node
	case NodeExtend:
		// p.floor.().PipeNode = node
	case NodeInclude:
		// p.floor.(*SetNode).PipeNode = node
	case NodeRange:
		p.floor.(*RangeNode).PipeNode = node
	}
}
