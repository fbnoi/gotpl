package template

type Parser struct {
	nodeStack []Node
	current   Node
	tokens    []*Token
	cursor    int
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
		case TYPE_BLOCK_END:
		case TYPE_VAR_END:
		case TYPE_NAME, TYPE_NUMBER, TYPE_STRING, TYPE_OPERATOR, TYPE_PUNCTUATION:
			p.parsePipeNode()
		default:
		}
	}
	return doc
}

func (p *Parser) parseText()
func (p *Parser) parseBlock()
func (p *Parser) parsePipeNode()
func (p *Parser) parseIfElse()
func (p *Parser) parseFor()
func (p *Parser) parseSet()
func (p *Parser) parseExtend()
func (p *Parser) parseInclude()
