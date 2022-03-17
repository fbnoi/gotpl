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
	NodeDoc                      // A document
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

const NoPos Pos = 0

type Pos int

func (p Pos) Position() Pos {
	return p
}

type Expression string

func (e Expression) Express() Expression {
	return e
}

type HasExpression interface {
	Express()
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
	// writeTo writes the String output to the builder.
	writeTo(*strings.Builder)
}

type TextNode struct {
	Pos
	NodeType
	Text string
}

func newTextNode(content string, pos int) *TextNode {
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

func (t *TextNode) writeTo(sb *strings.Builder) {
	sb.WriteString(t.String())
}

type ValueNode struct {
	Pos
	NodeType
	Expression
}

func newValueNode(pos int) *ValueNode {
	return &ValueNode{
		Pos:      Pos(pos),
		NodeType: NodeValue,
	}
}

func (v *ValueNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{{ ")
	sb.WriteString(string(v.Express()))
	sb.WriteString(" }}")
	return sb.String()
}

func (v *ValueNode) Copy() Node {
	return &ValueNode{
		Pos:        v.Pos,
		NodeType:   v.NodeType,
		Expression: v.Expression,
	}
}

func (v *ValueNode) writeTo(sb *strings.Builder) {
	sb.WriteString(v.String())
}

type IfNode struct {
	Pos
	NodeType
	Expression
	Branches []Node
}

func newIfNode(pos int) *IfNode {
	return &IfNode{
		Pos:      Pos(pos),
		NodeType: NodeIf,
	}
}

func (i *IfNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{{ if ")
	sb.WriteString(string(i.Express()))
	sb.WriteString(" }}")
	i.Walk(func(n Node) {
		sb.WriteString(n.String())
	})
	return sb.String()
}

func (i *IfNode) Copy() Node {
	return &IfNode{
		Pos:        i.Pos,
		NodeType:   i.NodeType,
		Branches:   copy(i.Branches),
		Expression: i.Expression,
	}
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
	Expression
	Branches []Node
}

func newElseIfNode(pos int) *ElseIfNode {
	return &ElseIfNode{
		Pos:      Pos(pos),
		NodeType: NodeIf,
	}
}

func (e *ElseIfNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{{ else if ")
	sb.WriteString(string(e.Express()))
	sb.WriteString(" }}")
	e.Walk(func(n Node) {
		sb.WriteString(n.String())
	})
	return sb.String()
}

func (e *ElseIfNode) Copy() Node {
	return &ElseIfNode{
		Pos:        e.Pos,
		NodeType:   e.NodeType,
		Expression: e.Expression,
		Branches:   copy(e.Branches),
	}
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

func newElseNode(pos int) *ElseNode {
	return &ElseNode{
		Pos:      Pos(pos),
		NodeType: NodeElse,
	}
}

func (e *ElseNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{{ else }}")
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

func newEndIfNode(pos int) *EndIfNode {
	return &EndIfNode{
		Pos:      Pos(pos),
		NodeType: NodeEndIf,
	}
}

func (e *EndIfNode) String() string {
	return "{{ end }}"
}

func (e *EndIfNode) Copy() Node {
	return &EndIfNode{
		Pos:      e.Pos,
		NodeType: e.NodeType,
	}
}

func (e *EndIfNode) writeTo(sb *strings.Builder) {
	sb.WriteString(e.String())
}

type RangeNode struct {
	Pos
	NodeType
	Expression
	Branches []Node
}

func newRangeNode(pos int) *RangeNode {
	return &RangeNode{
		Pos:      Pos(pos),
		NodeType: NodeRange,
	}
}

func (r *RangeNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{{ for ")
	sb.WriteString(string(r.Express()))
	sb.WriteString(" }}")
	r.Walk(func(n Node) {
		sb.WriteString(n.String())
	})
	return sb.String()
}

func (r *RangeNode) Copy() Node {
	return &RangeNode{
		Pos:        r.Pos,
		NodeType:   r.NodeType,
		Expression: r.Expression,
		Branches:   copy(r.Branches)}
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

func newEndRangeNode(pos int) *EndRangeNode {
	return &EndRangeNode{
		Pos:      Pos(pos),
		NodeType: NodeEndRange,
	}
}

func (e *EndRangeNode) String() string {
	return "{{ end }}"
}

func (e *EndRangeNode) Copy() Node {
	return &EndRangeNode{
		Pos:      e.Pos,
		NodeType: e.NodeType,
	}
}

func (e *EndRangeNode) writeTo(sb *strings.Builder) {
	sb.WriteString(e.String())
}

type SetNode struct {
	Pos
	NodeType
	Expression
}

func newSetNode(pos int) *SetNode {
	return &SetNode{Pos: Pos(pos)}
}

func (s *SetNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{{ Set ")
	sb.WriteString(string(s.Express()))
	sb.WriteString(" }}")
	return sb.String()
}

func (s *SetNode) Copy() Node {
	return &SetNode{
		Pos:        s.Pos,
		NodeType:   s.NodeType,
		Expression: s.Expression,
	}
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

func newBlockNode(name string, pos int) *BlockNode {
	return &BlockNode{
		Pos:      Pos(pos),
		NodeType: NodeBlock,
		Name:     name,
	}
}

func (b *BlockNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{{ block ")
	sb.WriteString(b.Name)
	sb.WriteString(" }}")
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
	return "{{ end }}"
}

func (e *EndBlockNode) Copy() Node {
	return &EndBlockNode{
		Pos:      e.Pos,
		NodeType: e.NodeType,
	}
}

func (e *EndBlockNode) writeTo(sb *strings.Builder) {
	sb.WriteString(e.String())
}

type ExtendNode struct {
	Pos
	NodeType
	Name string
}

func newExtendNode(name string, pos int) *ExtendNode {
	return &ExtendNode{
		Pos:      Pos(pos),
		NodeType: NodeExtend,
		Name:     name,
	}
}

func (b *ExtendNode) String() string {
	sb := &strings.Builder{}
	sb.WriteString("{{ extend ")
	sb.WriteString(b.Name)
	sb.WriteString(" }}")
	return sb.String()
}

func (b *ExtendNode) Copy() Node {
	return &ExtendNode{
		Pos:      b.Pos,
		NodeType: b.NodeType,
		Name:     b.Name,
	}
}

func (b *ExtendNode) writeTo(sb *strings.Builder) {
	sb.WriteString(b.String())
}

type ImportNode struct {
	Pos
	NodeType
}

func (i *ImportNode) String() string {
	return "{{ end }}"
}

func (i *ImportNode) Copy() Node {
	return &ImportNode{
		Pos:      i.Pos,
		NodeType: i.NodeType,
	}
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
