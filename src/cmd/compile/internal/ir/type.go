// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ir

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/types"
	"cmd/internal/src"
	"fmt"
)

// Nodes that represent the syntax of a type before type-checking.
// After type-checking, they serve only as shells around a *types.Type.
// Calling TypeNode converts a *types.Type to a Node shell.

// An Ntype is a Node that syntactically looks like a type.
// It can be the raw syntax for a type before typechecking,
// or it can be an OTYPE with Type() set to a *types.Type.
// Note that syntax doesn't guarantee it's a type: an expression
// like *fmt is an Ntype (we don't know whether names are types yet),
// but at least 1+1 is not an Ntype.
type Ntype interface {
	Node
	CanBeNtype()
}

// A miniType is a minimal type syntax Node implementation,
// to be embedded as the first field in a larger node implementation.
type miniType struct {
	miniNode
	typ *types.Type
}

func (*miniType) CanBeNtype() {}

func (n *miniType) Type() *types.Type { return n.typ }

// setOTYPE changes n to be an OTYPE node returning t.
// Rewriting the node in place this way should not be strictly
// necessary (we should be able to update the uses with
// proper OTYPE nodes), but it's mostly harmless and easy
// to keep doing for now.
//
// setOTYPE also records t.Nod = self if t.Nod is not already set.
// (Some types are shared by multiple OTYPE nodes, so only
// the first such node is used as t.Nod.)
func (n *miniType) setOTYPE(t *types.Type, self Node) {
	if n.typ != nil {
		panic(n.op.String() + " SetType: type already set")
	}
	n.op = OTYPE
	n.typ = t
	t.SetNod(self)
}

func (n *miniType) Sym() *types.Sym { return nil }   // for Format OTYPE
func (n *miniType) Implicit() bool  { return false } // for Format OTYPE

// A ChanType represents a chan Elem syntax with the direction Dir.
type ChanType struct {
	miniType
	Elem Node
	Dir  types.ChanDir
}

func NewChanType(pos src.XPos, elem Node, dir types.ChanDir) *ChanType {
	n := &ChanType{Elem: elem, Dir: dir}
	n.op = OTCHAN
	n.pos = pos
	return n
}

func (n *ChanType) String() string                { return fmt.Sprint(n) }
func (n *ChanType) Format(s fmt.State, verb rune) { FmtNode(n, s, verb) }
func (n *ChanType) copy() Node                    { c := *n; return &c }
func (n *ChanType) doChildren(do func(Node) error) error {
	var err error
	err = maybeDo(n.Elem, err, do)
	return err
}
func (n *ChanType) editChildren(edit func(Node) Node) {
	n.Elem = maybeEdit(n.Elem, edit)
}
func (n *ChanType) SetOTYPE(t *types.Type) {
	n.setOTYPE(t, n)
	n.Elem = nil
}

func (n *ChanType) DeepCopy(pos src.XPos) Node {
	if n.op == OTYPE {
		// Can't change types and no node references left.
		return n
	}
	return NewChanType(n.posOr(pos), DeepCopy(pos, n.Elem), n.Dir)
}

// A MapType represents a map[Key]Value type syntax.
type MapType struct {
	miniType
	Key  Node
	Elem Node
}

func NewMapType(pos src.XPos, key, elem Node) *MapType {
	n := &MapType{Key: key, Elem: elem}
	n.op = OTMAP
	n.pos = pos
	return n
}

func (n *MapType) String() string                { return fmt.Sprint(n) }
func (n *MapType) Format(s fmt.State, verb rune) { FmtNode(n, s, verb) }
func (n *MapType) copy() Node                    { c := *n; return &c }
func (n *MapType) doChildren(do func(Node) error) error {
	var err error
	err = maybeDo(n.Key, err, do)
	err = maybeDo(n.Elem, err, do)
	return err
}
func (n *MapType) editChildren(edit func(Node) Node) {
	n.Key = maybeEdit(n.Key, edit)
	n.Elem = maybeEdit(n.Elem, edit)
}
func (n *MapType) SetOTYPE(t *types.Type) {
	n.setOTYPE(t, n)
	n.Key = nil
	n.Elem = nil
}

func (n *MapType) DeepCopy(pos src.XPos) Node {
	if n.op == OTYPE {
		// Can't change types and no node references left.
		return n
	}
	return NewMapType(n.posOr(pos), DeepCopy(pos, n.Key), DeepCopy(pos, n.Elem))
}

// A StructType represents a struct { ... } type syntax.
type StructType struct {
	miniType
	Fields []*Field
}

func NewStructType(pos src.XPos, fields []*Field) *StructType {
	n := &StructType{Fields: fields}
	n.op = OTSTRUCT
	n.pos = pos
	return n
}

func (n *StructType) String() string                { return fmt.Sprint(n) }
func (n *StructType) Format(s fmt.State, verb rune) { FmtNode(n, s, verb) }
func (n *StructType) copy() Node {
	c := *n
	c.Fields = copyFields(c.Fields)
	return &c
}
func (n *StructType) doChildren(do func(Node) error) error {
	var err error
	err = maybeDoFields(n.Fields, err, do)
	return err
}
func (n *StructType) editChildren(edit func(Node) Node) {
	editFields(n.Fields, edit)
}

func (n *StructType) SetOTYPE(t *types.Type) {
	n.setOTYPE(t, n)
	n.Fields = nil
}

func (n *StructType) DeepCopy(pos src.XPos) Node {
	if n.op == OTYPE {
		// Can't change types and no node references left.
		return n
	}
	return NewStructType(n.posOr(pos), deepCopyFields(pos, n.Fields))
}

func deepCopyFields(pos src.XPos, fields []*Field) []*Field {
	var out []*Field
	for _, f := range fields {
		out = append(out, f.deepCopy(pos))
	}
	return out
}

// An InterfaceType represents a struct { ... } type syntax.
type InterfaceType struct {
	miniType
	Methods []*Field
}

func NewInterfaceType(pos src.XPos, methods []*Field) *InterfaceType {
	n := &InterfaceType{Methods: methods}
	n.op = OTINTER
	n.pos = pos
	return n
}

func (n *InterfaceType) String() string                { return fmt.Sprint(n) }
func (n *InterfaceType) Format(s fmt.State, verb rune) { FmtNode(n, s, verb) }
func (n *InterfaceType) copy() Node {
	c := *n
	c.Methods = copyFields(c.Methods)
	return &c
}
func (n *InterfaceType) doChildren(do func(Node) error) error {
	var err error
	err = maybeDoFields(n.Methods, err, do)
	return err
}
func (n *InterfaceType) editChildren(edit func(Node) Node) {
	editFields(n.Methods, edit)
}

func (n *InterfaceType) SetOTYPE(t *types.Type) {
	n.setOTYPE(t, n)
	n.Methods = nil
}

func (n *InterfaceType) DeepCopy(pos src.XPos) Node {
	if n.op == OTYPE {
		// Can't change types and no node references left.
		return n
	}
	return NewInterfaceType(n.posOr(pos), deepCopyFields(pos, n.Methods))
}

// A FuncType represents a func(Args) Results type syntax.
type FuncType struct {
	miniType
	Recv    *Field
	Params  []*Field
	Results []*Field
}

func NewFuncType(pos src.XPos, rcvr *Field, args, results []*Field) *FuncType {
	n := &FuncType{Recv: rcvr, Params: args, Results: results}
	n.op = OTFUNC
	n.pos = pos
	return n
}

func (n *FuncType) String() string                { return fmt.Sprint(n) }
func (n *FuncType) Format(s fmt.State, verb rune) { FmtNode(n, s, verb) }
func (n *FuncType) copy() Node {
	c := *n
	if c.Recv != nil {
		c.Recv = c.Recv.copy()
	}
	c.Params = copyFields(c.Params)
	c.Results = copyFields(c.Results)
	return &c
}
func (n *FuncType) doChildren(do func(Node) error) error {
	var err error
	err = maybeDoField(n.Recv, err, do)
	err = maybeDoFields(n.Params, err, do)
	err = maybeDoFields(n.Results, err, do)
	return err
}
func (n *FuncType) editChildren(edit func(Node) Node) {
	editField(n.Recv, edit)
	editFields(n.Params, edit)
	editFields(n.Results, edit)
}

func (n *FuncType) SetOTYPE(t *types.Type) {
	n.setOTYPE(t, n)
	n.Recv = nil
	n.Params = nil
	n.Results = nil
}

func (n *FuncType) DeepCopy(pos src.XPos) Node {
	if n.op == OTYPE {
		// Can't change types and no node references left.
		return n
	}
	return NewFuncType(n.posOr(pos),
		n.Recv.deepCopy(pos),
		deepCopyFields(pos, n.Params),
		deepCopyFields(pos, n.Results))
}

// A Field is a declared struct field, interface method, or function argument.
// It is not a Node.
type Field struct {
	Pos      src.XPos
	Sym      *types.Sym
	Ntype    Ntype
	Type     *types.Type
	Embedded bool
	IsDDD    bool
	Note     string
	Decl     *Name
}

func NewField(pos src.XPos, sym *types.Sym, ntyp Ntype, typ *types.Type) *Field {
	return &Field{Pos: pos, Sym: sym, Ntype: ntyp, Type: typ}
}

func (f *Field) String() string {
	var typ string
	if f.Type != nil {
		typ = fmt.Sprint(f.Type)
	} else {
		typ = fmt.Sprint(f.Ntype)
	}
	if f.Sym != nil {
		return fmt.Sprintf("%v %v", f.Sym, typ)
	}
	return typ
}

func (f *Field) copy() *Field {
	c := *f
	return &c
}

func copyFields(list []*Field) []*Field {
	out := make([]*Field, len(list))
	copy(out, list)
	for i, f := range out {
		out[i] = f.copy()
	}
	return out
}

func maybeDoField(f *Field, err error, do func(Node) error) error {
	if f != nil {
		if err == nil && f.Decl != nil {
			err = do(f.Decl)
		}
		if err == nil && f.Ntype != nil {
			err = do(f.Ntype)
		}
	}
	return err
}

func maybeDoFields(list []*Field, err error, do func(Node) error) error {
	if err != nil {
		return err
	}
	for _, f := range list {
		err = maybeDoField(f, err, do)
		if err != nil {
			return err
		}
	}
	return err
}

func editField(f *Field, edit func(Node) Node) {
	if f == nil {
		return
	}
	if f.Decl != nil {
		f.Decl = edit(f.Decl).(*Name)
	}
	if f.Ntype != nil {
		f.Ntype = toNtype(edit(f.Ntype))
	}
}

func editFields(list []*Field, edit func(Node) Node) {
	for _, f := range list {
		editField(f, edit)
	}
}

func (f *Field) deepCopy(pos src.XPos) *Field {
	if f == nil {
		return nil
	}
	fpos := pos
	if !pos.IsKnown() {
		fpos = f.Pos
	}
	decl := f.Decl
	if decl != nil {
		decl = DeepCopy(pos, decl).(*Name)
	}
	ntype := f.Ntype
	if ntype != nil {
		ntype = DeepCopy(pos, ntype).(Ntype)
	}
	// No keyed literal here: if a new struct field is added, we want this to stop compiling.
	return &Field{fpos, f.Sym, ntype, f.Type, f.Embedded, f.IsDDD, f.Note, decl}
}

// A SliceType represents a []Elem type syntax.
// If DDD is true, it's the ...Elem at the end of a function list.
type SliceType struct {
	miniType
	Elem Node
	DDD  bool
}

func NewSliceType(pos src.XPos, elem Node) *SliceType {
	n := &SliceType{Elem: elem}
	n.op = OTSLICE
	n.pos = pos
	return n
}

func (n *SliceType) String() string                { return fmt.Sprint(n) }
func (n *SliceType) Format(s fmt.State, verb rune) { FmtNode(n, s, verb) }
func (n *SliceType) copy() Node                    { c := *n; return &c }
func (n *SliceType) doChildren(do func(Node) error) error {
	var err error
	err = maybeDo(n.Elem, err, do)
	return err
}
func (n *SliceType) editChildren(edit func(Node) Node) {
	n.Elem = maybeEdit(n.Elem, edit)
}
func (n *SliceType) SetOTYPE(t *types.Type) {
	n.setOTYPE(t, n)
	n.Elem = nil
}

func (n *SliceType) DeepCopy(pos src.XPos) Node {
	if n.op == OTYPE {
		// Can't change types and no node references left.
		return n
	}
	return NewSliceType(n.posOr(pos), DeepCopy(pos, n.Elem))
}

// An ArrayType represents a [Len]Elem type syntax.
// If Len is nil, the type is a [...]Elem in an array literal.
type ArrayType struct {
	miniType
	Len  Node
	Elem Node
}

func NewArrayType(pos src.XPos, size Node, elem Node) *ArrayType {
	n := &ArrayType{Len: size, Elem: elem}
	n.op = OTARRAY
	n.pos = pos
	return n
}

func (n *ArrayType) String() string                { return fmt.Sprint(n) }
func (n *ArrayType) Format(s fmt.State, verb rune) { FmtNode(n, s, verb) }
func (n *ArrayType) copy() Node                    { c := *n; return &c }
func (n *ArrayType) doChildren(do func(Node) error) error {
	var err error
	err = maybeDo(n.Len, err, do)
	err = maybeDo(n.Elem, err, do)
	return err
}
func (n *ArrayType) editChildren(edit func(Node) Node) {
	n.Len = maybeEdit(n.Len, edit)
	n.Elem = maybeEdit(n.Elem, edit)
}

func (n *ArrayType) DeepCopy(pos src.XPos) Node {
	if n.op == OTYPE {
		// Can't change types and no node references left.
		return n
	}
	return NewArrayType(n.posOr(pos), DeepCopy(pos, n.Len), DeepCopy(pos, n.Elem))
}

func (n *ArrayType) SetOTYPE(t *types.Type) {
	n.setOTYPE(t, n)
	n.Len = nil
	n.Elem = nil
}

// A typeNode is a Node wrapper for type t.
type typeNode struct {
	miniNode
	typ *types.Type
}

func newTypeNode(pos src.XPos, typ *types.Type) *typeNode {
	n := &typeNode{typ: typ}
	n.pos = pos
	n.op = OTYPE
	return n
}

func (n *typeNode) String() string                { return fmt.Sprint(n) }
func (n *typeNode) Format(s fmt.State, verb rune) { FmtNode(n, s, verb) }
func (n *typeNode) copy() Node                    { c := *n; return &c }
func (n *typeNode) doChildren(do func(Node) error) error {
	return nil
}
func (n *typeNode) editChildren(edit func(Node) Node) {}

func (n *typeNode) Type() *types.Type { return n.typ }
func (n *typeNode) Sym() *types.Sym   { return n.typ.Sym() }
func (n *typeNode) CanBeNtype()       {}

// TypeNode returns the Node representing the type t.
func TypeNode(t *types.Type) Ntype {
	if n := t.Obj(); n != nil {
		if n.Type() != t {
			base.Fatalf("type skew: %v has type %v, but expected %v", n, n.Type(), t)
		}
		return n.(Ntype)
	}
	return newTypeNode(src.NoXPos, t)
}