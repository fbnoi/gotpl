package template

import (
	"strings"
)

const (
	FnMethod FuncType = iota
	FnFunc
)

const (
	ArgField ArgType = iota
	ArgFunc
)

const (
	NodeText     NodeType = iota // Plain text.
	NodeValue                    // A non-control action such as a field evaluation.
	NodePipe                     // A non-control action such as a field evaluation.
	NodeIf                       // An if action.
	NodeElif                     // An else if action
	NodeElse                     // An else action
	NodeEndIf                    // An end if action
	NodeRange                    // A range action.
	NodeEndRange                 // An end range action.
	NodeBranch                   // A branch node, maybe a range branch or a if branch
	NodeSet                      // A set action
	NodeBlock                    // A block action
	NodeEndBlock                 // An end block action
	NodeImport                   // An import action
	NodeDoc                      // An document node
	NodeExtend                   // An extend action
	NodeInclude                  // An include action
)

// NodeType identifies the type of a parse tree node.
type NodeType int

// Type returns itself and provides an easy default implementation
// for embedding in a Node. Embedded in all non-trivial Nodes.
func (t NodeType) Type() NodeType {
	return t
}

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos int

func (p Pos) Position() Pos {
	return p
}

type FuncType int

type FuncInfo struct {
	Signature string
	Typ       FuncType
}

type ArgType int

type ArgInfo struct {
	Name string
	Typ  ArgType
}

type Node interface {
	Type() NodeType
	String() string
	// Copy does a deep copy of the Node and all its components.
	// To avoid type assertions, some XxxNodes also have specialized
	// CopyXxx methods that return *XxxNode.
	Copy() Node
	Position() Pos // byte position of start of node in full original input string
	// It is unexported so all implementations of Node are in this package.
	document() *Document
	// writeTo writes the String output to the builder.
	writeTo(*strings.Builder)
}

type TextNode struct {
	Pos
	NodeType
	doc  *Document
	Text []byte
}

func (t *TextNode) String() string {
	return string(t.Text)
}

func (t *TextNode) Copy() Node {
	return &TextNode{Pos: t.Pos, NodeType: t.NodeType, Text: t.Text, doc: t.doc}
}

func (t *TextNode) document() *Document {
	return t.doc
}

func (t *TextNode) writeTo(sb *strings.Builder) {
	sb.WriteString(t.String())
}

type ValueNode struct {
	Pos
	NodeType
	doc      *Document
	PipeNode *PipeNode
}

func (v *ValueNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{{ ")
	sb.WriteString(v.PipeNode.String())
	sb.WriteString(" }}")
	return sb.String()
}

func (v *ValueNode) Copy() Node {
	return &ValueNode{
		Pos:      v.Pos,
		NodeType: v.NodeType,
		doc:      v.doc,
		PipeNode: v.PipeNode.Copy().(*PipeNode),
	}
}

func (v *ValueNode) document() *Document {
	return v.doc
}

func (v *ValueNode) writeTo(sb *strings.Builder) {
	sb.WriteString(v.String())
}

type PipeNode struct {
	Pos
	NodeType
	doc      *Document
	FnStack  []FuncInfo
	ArgStack [][]ArgInfo
}

func (p *PipeNode) String() string {
	return ""
}

func (p *PipeNode) Copy() Node {
	return &PipeNode{
		Pos:      p.Pos,
		NodeType: p.NodeType,
		doc:      p.doc,
		FnStack:  p.FnStack,
		ArgStack: p.ArgStack,
	}
}

func (p *PipeNode) document() *Document {
	return p.doc
}

func (p *PipeNode) writeTo(sb *strings.Builder) {
	sb.WriteString(p.String())
}

type IfNode struct {
	Pos
	NodeType
	doc      *Document
	PipeNode *PipeNode
	Branchs  []Node
}

func (i *IfNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{% if")
	sb.WriteString(i.PipeNode.String())
	sb.WriteString(" %}")
	for _, node := range i.Branchs {
		sb.WriteString(node.String())
	}
	return sb.String()
}

func (i *IfNode) Copy() Node {
	return &IfNode{Pos: i.Pos, NodeType: i.NodeType, PipeNode: i.PipeNode.Copy().(*PipeNode)}
}

func (i *IfNode) document() *Document {
	return i.doc
}

func (i *IfNode) writeTo(sb *strings.Builder) {
	sb.WriteString(i.String())
}

type ElseIfNode struct {
	Pos
	NodeType
	doc      *Document
	PipeNode *PipeNode
	Branchs  []Node
}

func (e *ElseIfNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{% elseif")
	sb.WriteString(e.PipeNode.String())
	sb.WriteString(" %}")
	for _, node := range e.Branchs {
		sb.WriteString(node.String())
	}
	return sb.String()
}

func (e *ElseIfNode) Copy() Node {
	return &ElseIfNode{Pos: e.Pos, NodeType: e.NodeType, PipeNode: e.PipeNode.Copy().(*PipeNode), doc: e.doc}
}

func (e *ElseIfNode) document() *Document {
	return e.doc
}

func (e *ElseIfNode) writeTo(sb *strings.Builder) {
	sb.WriteString(e.String())
}

type ElseNode struct {
	Pos
	NodeType
	doc     *Document
	Branchs []Node
}

func (e *ElseNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{% else %}")
	for _, node := range e.Branchs {
		sb.WriteString(node.String())
	}
	return sb.String()
}

func (e *ElseNode) Copy() Node {
	return &ElseNode{Pos: e.Pos, NodeType: e.NodeType, doc: e.doc}
}

func (e *ElseNode) document() *Document {
	return e.doc
}

func (e *ElseNode) writeTo(sb *strings.Builder) {
	sb.WriteString(e.String())
}

type EndIfNode struct {
	Pos
	NodeType
	doc *Document
}

func (e *EndIfNode) String() string {
	return "{% endif %}"
}

func (e *EndIfNode) Copy() Node {
	return &EndIfNode{Pos: e.Pos, NodeType: e.NodeType, doc: e.doc}
}

func (e *EndIfNode) document() *Document {
	return e.doc
}

func (e *EndIfNode) writeTo(sb *strings.Builder) {
	sb.WriteString(e.String())
}

type RangeNode struct {
	Pos
	NodeType
	doc      *Document
	PipeNode *PipeNode
	Branchs  []Node
}

func (r *RangeNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{% for ")
	sb.WriteString(r.PipeNode.String())
	sb.WriteString(" %}")
	for _, node := range r.Branchs {
		sb.WriteString(node.String())
	}
	return sb.String()
}

func (r *RangeNode) Copy() Node {
	return &RangeNode{Pos: r.Pos, NodeType: r.NodeType, PipeNode: r.PipeNode.Copy().(*PipeNode), doc: r.doc}
}

func (r *RangeNode) document() *Document {
	return r.doc
}

func (r *RangeNode) writeTo(sb *strings.Builder) {
	sb.WriteString(r.String())
}

type EndRangeNode struct {
	Pos
	NodeType
	doc      *Document
	PipeNode *PipeNode
	Branchs  []Node
}

func (e *EndRangeNode) String() string {
	return "{% endfor %}"
}

func (e *EndRangeNode) Copy() Node {
	return &EndRangeNode{Pos: e.Pos, NodeType: e.NodeType, doc: e.doc}
}

func (e *EndRangeNode) document() *Document {
	return e.doc
}

func (e *EndRangeNode) writeTo(sb *strings.Builder) {
	sb.WriteString(e.String())
}

type SetNode struct {
	Pos
	NodeType
	doc      *Document
	PipeNode *PipeNode
}

func (s *SetNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{% Set ")
	sb.WriteString(s.PipeNode.String())
	sb.WriteString(" %}")
	return sb.String()
}

func (s *SetNode) Copy() Node {
	return &SetNode{Pos: s.Pos, NodeType: s.NodeType, PipeNode: s.PipeNode, doc: s.doc}
}

func (s *SetNode) document() *Document {
	return s.doc
}

func (s *SetNode) writeTo(sb *strings.Builder) {
	sb.WriteString(s.String())
}

type BlockNode struct {
	Pos
	NodeType
	doc     *Document
	Name    string
	Branchs []Node
}

func (b *BlockNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{% block ")
	sb.WriteString(b.Name)
	sb.WriteString(" %}")
	for _, n := range b.Branchs {
		sb.WriteString(n.String())
	}
	return sb.String()
}

func (b *BlockNode) Copy() Node {
	return &BlockNode{Pos: b.Pos, NodeType: b.NodeType, Branchs: b.Branchs, doc: b.doc}
}

func (b *BlockNode) document() *Document {
	return b.doc
}

func (b *BlockNode) writeTo(sb *strings.Builder) {
	sb.WriteString(b.String())
}

type EndBlockNode struct {
	Pos
	NodeType
	doc *Document
}

func (e *EndBlockNode) String() string {
	return "{% endblock %}"
}

func (e *EndBlockNode) Copy() Node {
	return &EndBlockNode{Pos: e.Pos, NodeType: e.NodeType, doc: e.doc}
}

func (b *EndBlockNode) document() *Document {
	return b.doc
}

func (e *EndBlockNode) writeTo(sb *strings.Builder) {
	sb.WriteString(e.String())
}

type ImportNode struct {
	Pos
	NodeType
	doc *Document
}

func (i *ImportNode) String() string {
	return "{% endblock %}"
}

func (i *ImportNode) Copy() Node {
	return &ImportNode{Pos: i.Pos, NodeType: i.NodeType, doc: i.doc}
}

func (b *ImportNode) document() *Document {
	return b.doc
}

func (i *ImportNode) writeTo(sb *strings.Builder) {
	sb.WriteString(i.String())
}
