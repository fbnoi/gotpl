package template

import (
	"strings"
)

const (
	NodeText     NodeType = iota // Plain text.
	NodeValue                    // A non-control action such as a field evaluation.
	NodePipe                     // A non-control action such as a field evaluation.
	NodeIf                       // An if action.
	NodeElseif                   // An else if action
	NodeElse                     // An else action
	NodeEndIf                    // An end if action
	NodeRange                    // A range action.
	NodeEndRange                 // An end range action.
	NodeBranch                   // A branch node, maybe a range branch or a if branch
	NodeSet                      // A set action
	NodeBlock                    // A block action
	NodeEndBlock                 // An end block action
	NodeImport                   // An import action
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

type BranchAble interface {
	Append(Node)
	Walk(func(Node))
}

type Node interface {
	Type() NodeType
	String() string
	// Copy does a deep copy of the Node and all its components.
	// To avoid type assertions, some XxxNodes also have specialized
	// CopyXxx methods that return *XxxNode.
	Copy() Node
	Position() Pos // byte position of start of node in full original input string
	Parse() string
	// writeTo writes the String output to the builder.
	writeTo(*strings.Builder)
}

type TextNode struct {
	Pos
	NodeType
	Text string
}

func newTextNode(content string, pos int, doc *Document) *TextNode {
	return &TextNode{
		Pos:      Pos(pos),
		NodeType: NodeText,
	}
}

func (t *TextNode) String() string {
	return string(t.Text)
}

func (t *TextNode) Copy() Node {
	return &TextNode{
		Pos:      t.Pos,
		NodeType: t.NodeType,
		Text:     t.Text,
	}
}

func (t *TextNode) Parse() string {
	return t.String()
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

func (v *ValueNode) Parse() string {
	sb := &strings.Builder{}
	sb.WriteString("{{")
	sb.WriteString(v.PipeNode.String())
	sb.WriteString("}}")
	return sb.String()
}

func (v *ValueNode) writeTo(sb *strings.Builder) {
	sb.WriteString(v.String())
}

type PipeNode struct {
	Pos
	NodeType
	express string
}

func newPipeNode(express string, pos int) *PipeNode {
	return &PipeNode{
		Pos:      Pos(pos),
		NodeType: NodePipe,
		express:  express,
	}
}

func (p *PipeNode) String() string {
	return p.express
}

func (p *PipeNode) Copy() Node {
	return p.CopyPipeNode()
}
func (p *PipeNode) CopyPipeNode() *PipeNode {
	return &PipeNode{
		Pos:      p.Pos,
		NodeType: p.NodeType,
		express:  p.express,
	}
}

func (p *PipeNode) Parse() string {
	return p.String()
}

func (p *PipeNode) writeTo(sb *strings.Builder) {
	sb.WriteString(p.String())
}

type IfNode struct {
	Pos
	NodeType
	PipeNode *PipeNode
	Branches []Node
}

func newIfNode(pos int, doc *Document) *IfNode {
	return &IfNode{
		Pos:      Pos(pos),
		NodeType: NodeIf,
	}
}

func (i *IfNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{% if ")
	sb.WriteString(i.PipeNode.String())
	sb.WriteString(" %}")
	i.Walk(func(n Node) {
		sb.WriteString(n.String())
	})
	return sb.String()
}

func (i *IfNode) Copy() Node {
	return &IfNode{
		Pos:      i.Pos,
		NodeType: i.NodeType,
		PipeNode: i.PipeNode.CopyPipeNode(),
		Branches: copy(i.Branches),
	}
}

func (i *IfNode) Parse() string {
	sb := &strings.Builder{}
	sb.WriteString("{{ if ")
	sb.WriteString(i.PipeNode.String())
	sb.WriteString(" }}")
	i.Walk(func(n Node) {
		sb.WriteString(n.String())
	})
	return sb.String()
}

func (i *IfNode) Walk(fn func(n Node)) {
	for _, n := range i.Branches {
		fn(n)
	}
}

func (i *IfNode) Append(n Node) {
	i.Branches = append(i.Branches, n)
}

func (i *IfNode) writeTo(sb *strings.Builder) {
	sb.WriteString(i.String())
}

type ElseIfNode struct {
	Pos
	NodeType
	PipeNode *PipeNode
	Branches []Node
}

func newElseIfNode(pos int, doc *Document) *ElseIfNode {
	return &ElseIfNode{
		Pos:      Pos(pos),
		NodeType: NodeIf,
	}
}

func (e *ElseIfNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{% elseif ")
	sb.WriteString(e.PipeNode.String())
	sb.WriteString(" %}")
	e.Walk(func(n Node) {
		sb.WriteString(n.String())
	})
	return sb.String()
}

func (e *ElseIfNode) Copy() Node {
	return &ElseIfNode{
		Pos:      e.Pos,
		NodeType: e.NodeType,
		PipeNode: e.PipeNode.CopyPipeNode(),
		Branches: copy(e.Branches),
	}
}

func (e *ElseIfNode) Parse() string {
	sb := &strings.Builder{}
	sb.WriteString("{{ else if ")
	sb.WriteString(e.PipeNode.String())
	sb.WriteString(" }}")
	e.Walk(func(n Node) {
		sb.WriteString(n.String())
	})
	return sb.String()
}

func (e *ElseIfNode) Walk(fn func(n Node)) {
	for _, n := range e.Branches {
		fn(n)
	}
}

func (e *ElseIfNode) Append(n Node) {
	e.Branches = append(e.Branches, n)
}

func (e *ElseIfNode) writeTo(sb *strings.Builder) {
	sb.WriteString(e.String())
}

type ElseNode struct {
	Pos
	NodeType
	Branches []Node
}

func (e *ElseNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{% else %}")
	e.Walk(func(n Node) {
		sb.WriteString(n.String())
	})
	return sb.String()
}

func (e *ElseNode) Copy() Node {
	return &ElseNode{
		Pos:      e.Pos,
		NodeType: e.NodeType,
		Branches: copy(e.Branches),
	}
}

func (e *ElseNode) Parse() string {
	sb := &strings.Builder{}
	sb.WriteString("{{ else }}")
	e.Walk(func(n Node) {
		sb.WriteString(n.String())
	})
	return sb.String()
}

func (e *ElseNode) Walk(fn func(n Node)) {
	for _, n := range e.Branches {
		fn(n)
	}
}

func (e *ElseNode) Append(n Node) {
	e.Branches = append(e.Branches, n)
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
	return &EndIfNode{
		Pos:      e.Pos,
		NodeType: e.NodeType,
	}
}

func (e *EndIfNode) Parse() string {
	return "{{ end }}"
}

func (e *EndIfNode) writeTo(sb *strings.Builder) {
	sb.WriteString(e.String())
}

type RangeNode struct {
	Pos
	NodeType
	PipeNode *PipeNode
	Branches []Node
}

func (r *RangeNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{% for ")
	sb.WriteString(r.PipeNode.String())
	sb.WriteString(" %}")
	r.Walk(func(n Node) {
		sb.WriteString(n.String())
	})
	return sb.String()
}

func (r *RangeNode) Copy() Node {
	return &RangeNode{
		Pos:      r.Pos,
		NodeType: r.NodeType,
		PipeNode: r.PipeNode.CopyPipeNode(),
		Branches: copy(r.Branches)}
}

func (r *RangeNode) Parse() string {
	sb := &strings.Builder{}
	sb.WriteString("{{ for ")
	sb.WriteString(r.PipeNode.String())
	sb.WriteString(" }}")
	r.Walk(func(n Node) {
		sb.WriteString(n.String())
	})
	return sb.String()
}

func (e *RangeNode) Walk(fn func(n Node)) {
	for _, n := range e.Branches {
		fn(n)
	}
}

func (e *RangeNode) Append(n Node) {
	e.Branches = append(e.Branches, n)
}

func (r *RangeNode) writeTo(sb *strings.Builder) {
	sb.WriteString(r.String())
}

type EndRangeNode struct {
	Pos
	NodeType
}

func (e *EndRangeNode) String() string {
	return "{% endfor %}"
}

func (e *EndRangeNode) Copy() Node {
	return &EndRangeNode{
		Pos:      e.Pos,
		NodeType: e.NodeType,
	}
}

func (e *EndRangeNode) Parse() string {
	return "{{ end }}"
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
	return &SetNode{
		Pos:      s.Pos,
		NodeType: s.NodeType,
		PipeNode: s.PipeNode.CopyPipeNode(),
	}
}

func (s *SetNode) Parse() string {
	sb := &strings.Builder{}
	sb.WriteString("{{ Set ")
	sb.WriteString(s.PipeNode.String())
	sb.WriteString(" }}")
	return sb.String()
}

func (s *SetNode) writeTo(sb *strings.Builder) {
	sb.WriteString(s.String())
}

type BlockNode struct {
	Pos
	NodeType
	Name     string
	Branches []Node
}

func (b *BlockNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{% block ")
	sb.WriteString(b.Name)
	sb.WriteString(" %}")
	b.Walk(func(n Node) {
		sb.WriteString(n.String())
	})
	return sb.String()
}

func (b *BlockNode) Copy() Node {
	return &BlockNode{
		Pos:      b.Pos,
		NodeType: b.NodeType,
		Name:     b.Name,
		Branches: copy(b.Branches),
	}
}

func (b *BlockNode) Parse() string {
	sb := &strings.Builder{}
	sb.WriteString("{% block ")
	sb.WriteString(b.Name)
	sb.WriteString(" %}")
	b.Walk(func(n Node) {
		sb.WriteString(n.String())
	})
	return sb.String()
}

func (e *BlockNode) Walk(fn func(n Node)) {
	for _, n := range e.Branches {
		fn(n)
	}
}

func (e *BlockNode) Append(n Node) {
	e.Branches = append(e.Branches, n)
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
	return &EndBlockNode{
		Pos:      e.Pos,
		NodeType: e.NodeType,
	}
}

func (b *EndBlockNode) Parse() string {
	return "{{ end }}"
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
	return &ImportNode{
		Pos:      i.Pos,
		NodeType: i.NodeType,
	}
}

func (i *ImportNode) Parse() string {
	return "{{ end }}"
}

func (i *ImportNode) writeTo(sb *strings.Builder) {
	sb.WriteString(i.String())
}

func copy(ns []Node) []Node {
	nns := make([]Node, len(ns))
	for j := 0; j < len(ns); j++ {
		nns[j] = ns[j].Copy()
	}
	return nns
}
