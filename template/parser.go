package template

type BranchAbleNode interface {
	Node
	BranchAble
}

type Parser struct {
	branchStack []BranchAbleNode
	stream      *TokenStream
	upper       BranchAbleNode
	cursor      Node
	doc         *Document
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
		default:
		}
	}
	if p.upper != p.doc {
		panic("")
	}
	return p.doc
}

func (p *Parser) parseText(token *Token) {
	p.cursor = newTextNode(token.Value(), token.At)
	p.upper.Append(p.cursor)
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
			p.parseEndIfNode(token)
		case "set":
			p.parseSetNode(token)
		case "for":
			p.parseForNode(token)
		case "endfor":
			p.parseEndForNode(token)
		case "block":
			p.parseBlockNode(token)
		case "endblock":
			p.parseEndBlockNode(token)
		case "extend":
			p.parseExtendNode(token)
		case "include":
			p.parseIncludeNode(token)
		}
	}
}

func (p *Parser) parseIncludeNode(token *Token) {

}

func (p *Parser) parseExtendNode(token *Token) {
	var name string
	for token.Type() != TYPE_BLOCK_END {
		if name != "" {
			panic("")
		}
		name = token.Value()
	}
	node := newExtendNode(name, token.At)
	p.upper.Append(node)
	p.moveCursor(node)
}

func (p *Parser) parseEndBlockNode(token *Token) {
	if p.cursor.Type() == NodeEndBlock {
		node := newEndRangeNode(token.At)
		p.upper.Append(node)
		p.moveCursor(node)
	}
	panic("")
}

func (p *Parser) parseBlockNode(token *Token) {
	var name string
	for token.Type() != TYPE_BLOCK_END {
		if name != "" {
			panic("")
		}
		name = token.Value()
	}
	node := newBlockNode(name, token.At)
	p.upper.Append(node)
	p.moveCursor(node)
}

func (p *Parser) parseEndForNode(token *Token) {
	if p.cursor.Type() == NodeEndRange {
		node := newEndRangeNode(token.At)
		p.upper.Append(node)
		p.moveCursor(node)
	}
	panic("")
}

func (p *Parser) parseForNode(token *Token) {
	node := newRangeNode(token.At)
	p.evalExpression(node)
	p.upper.Append(node)
	p.moveCursor(node)
}

func (p *Parser) parseSetNode(token *Token) {
	node := newSetNode(token.At)
	p.evalExpression(node)
	p.upper.Append(node)
	p.moveCursor(node)
}

func (p *Parser) parseIfNode(token *Token) {
	node := newIfNode(token.At)
	p.evalExpression(node)
	p.upper.Append(node)
	p.moveCursor(node)
}

func (p *Parser) parseElseIfNode(token *Token) {
	if p.cursor.Type() == NodeIf || p.cursor.Type() == NodeElseif {
		node := newElseIfNode(token.At)
		p.evalExpression(node)
		p.upper.Append(node)
		p.moveCursor(node)
		p.upper = node
	}
	panic("")
}

func (p *Parser) parseElseNode(token *Token) {
	if p.cursor.Type() == NodeIf || p.cursor.Type() == NodeElseif {
		node := newElseNode(token.At)
		p.upper.Append(node)
		p.moveCursor(node)
		p.upper = node
	}
	panic("")
}

func (p *Parser) parseEndIfNode(token *Token) {
	if p.upper.Type() == NodeIf || p.cursor.Type() == NodeElseif {
		node := newEndIfNode(token.At)
		p.upper.Append(node)
		p.moveCursor(node)
	}
	panic("")
}

func (p *Parser) evalExpression(n Node) {
	// sb := &strings.Builder{}
	// for token.Type() != TYPE_BLOCK_END {
	// 	sb.WriteString(token.Value())
	// 	token = p.stream.Next()
	// }
	// node.Expression = Expression(sb.String())
	// p.cursor = node
}

func (p *Parser) pushStack(n BranchAbleNode) {
	p.branchStack = append(p.branchStack, p.upper)
	p.upper = n
}

func (p *Parser) popStack() BranchAbleNode {
	l := len(p.branchStack)
	if l == 0 {
		panic("")
	}

	p.upper = p.branchStack[l-1]
	p.branchStack = p.branchStack[:l-1]
	return p.upper
}

func (p *Parser) moveCursor(n Node) {
	p.cursor = n
	switch n.Type() {
	case NodeIf, NodeRange, NodeBlock:
		p.pushStack(n.(BranchAbleNode))
	case NodeElse, NodeElseif:
		p.upper = n.(BranchAbleNode)
	case NodeEndIf, NodeEndRange, NodeEndBlock:
		p.popStack()
	}
}
