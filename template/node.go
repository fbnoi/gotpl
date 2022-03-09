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
	// writeTo writes the String output to the builder.
	writeTo(*strings.Builder)
}

type TextNode struct {
	Pos
	NodeType
	Text []byte
}

func (t *TextNode) String() string {
	return string(t.Text)
}

func (t *TextNode) Copy() Node {
	return &TextNode{Pos: t.Pos, NodeType: t.NodeType, Text: t.Text}
}

func (t *TextNode) writeTo(sb *strings.Builder) {
	sb.WriteString(t.String())
}

type ValueNode struct {
	Pos
	NodeType
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
		PipeNode: v.PipeNode.Copy().(*PipeNode),
	}
}

func (v *ValueNode) writeTo(sb *strings.Builder) {
	sb.WriteString(v.String())
}

type PipeNode struct {
	Pos
	NodeType
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
		FnStack:  p.FnStack,
		ArgStack: p.ArgStack,
	}
}

func (p *PipeNode) writeTo(sb *strings.Builder) {
	sb.WriteString(p.String())
}

type IfNode struct {
	Pos
	NodeType
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

func (i *IfNode) writeTo(sb *strings.Builder) {
	sb.WriteString(i.String())
}

type ElseIfNode struct {
	Pos
	NodeType
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
	return &IfNode{Pos: e.Pos, NodeType: e.NodeType, PipeNode: e.PipeNode.Copy().(*PipeNode)}
}

func (e *ElseIfNode) writeTo(sb *strings.Builder) {
	sb.WriteString(e.String())
}

type ElseNode struct {
	Pos
	NodeType
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
	return &IfNode{Pos: e.Pos, NodeType: e.NodeType}
}

func (e *ElseNode) writeTo(sb *strings.Builder) {
	sb.WriteString(e.String())
}

type EndIfNode struct {
	Pos
	NodeType
}

func (e *EndIfNode) String() string {
	return "{% endif %}"
}

func (e *EndIfNode) Copy() Node {
	return &IfNode{Pos: e.Pos, NodeType: e.NodeType}
}

func (e *EndIfNode) writeTo(sb *strings.Builder) {
	sb.WriteString(e.String())
}

type RangeNode struct {
	Pos
	NodeType
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
	return &IfNode{Pos: r.Pos, NodeType: r.NodeType, PipeNode: r.PipeNode.Copy().(*PipeNode)}
}

func (r *RangeNode) writeTo(sb *strings.Builder) {
	sb.WriteString(r.String())
}

type EndRangeNode struct {
	Pos
	NodeType
	PipeNode *PipeNode
	Branchs  []Node
}

func (e *EndRangeNode) String() string {
	return "{% endfor %}"
}

func (e *EndRangeNode) Copy() Node {
	return &IfNode{Pos: e.Pos, NodeType: e.NodeType}
}

func (e *EndRangeNode) writeTo(sb *strings.Builder) {
	sb.WriteString(e.String())
}

type SetNode struct {
	Pos
	NodeType
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
	return &SetNode{Pos: s.Pos, NodeType: s.NodeType, PipeNode: s.PipeNode}
}

func (s *SetNode) writeTo(sb *strings.Builder) {
	sb.WriteString(s.String())
}

type BlockNode struct {
	Pos
	NodeType
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
	return &BlockNode{Pos: b.Pos, NodeType: b.NodeType, Branchs: b.Branchs}
}

func (b *BlockNode) writeTo(sb *strings.Builder) {
	sb.WriteString(b.String())
}

type EndBlockNode struct {
	Pos
	NodeType
}

func (e *EndBlockNode) String() string {
	return "{% endblock %}"
}

func (e *EndBlockNode) Copy() Node {
	return &EndBlockNode{Pos: e.Pos, NodeType: e.NodeType}
}

func (e *EndBlockNode) writeTo(sb *strings.Builder) {
	sb.WriteString(e.String())
}

type ImportNode struct {
	Pos
	NodeType
}

func (i *ImportNode) String() string {
	return "{% endblock %}"
}

func (i *ImportNode) Copy() Node {
	return &ImportNode{Pos: i.Pos, NodeType: i.NodeType}
}

func (i *ImportNode) writeTo(sb *strings.Builder) {
	sb.WriteString(i.String())
}
