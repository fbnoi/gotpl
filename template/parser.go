package template

type Parser struct {
	nodeState []Node
	spot      Node
	tokens    []*Token
	current   int
}

func (p *Parser) Parse(stream *TokenStream) *Document {
	doc := &Document{}
	for !stream.IsEOF() {
		token := stream.Next()
		switch token.Type() {
		case TYPE_TEXT:
			p.parseText()
		case TYPE_BLOCK_START:
			p.parseBlock()
		case TYPE_VAR_START:
			// p.parseBlock()
		case TYPE_BLOCK_END:
		case TYPE_VAR_END:
		case TYPE_NAME, TYPE_NUMBER, TYPE_STRING, TYPE_OPERATOR, TYPE_PUNCTUATION:
			p.parsePipeNode()
		default:
		}
	}
	return doc
}

func (p *Parser) parseText() *TextNode
func (p *Parser) parseBlock() Node
func (p *Parser) parsePipeNode() *TextNode
func (p *Parser) parseIfElse() Node
func (p *Parser) parseFor() *RangeNode
func (p *Parser) parseSet() *SetNode
func (p *Parser) parseExtend() Node
func (p *Parser) parseInclude() Node
