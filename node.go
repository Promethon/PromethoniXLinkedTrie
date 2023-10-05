package PromethoniXTrie

import (
	"bytes"
	"io"
)

type Data []byte
type Hash []byte
type Route Hash

type NodeType int8

const (
	_ NodeType = iota
	Extension
	Leaf
	Branch
)

type Node interface {
	Type() NodeType
	Encode(io.Writer) error
	Decode(io.Reader) error
	NextRoute(Route) (Hash, Route, error)

	Details() *NodeDetails
	EncodeAndHash() error
}

type NodeDetails struct {
	_node       Node
	Hash        Hash
	EncodedData Data
}

func EncodeNode(node Node) (Data, error) {
	buf := bytes.Buffer{}
	err := writeNodeType(&buf, node.Type())
	if err == nil {
		err = node.Encode(&buf)
	}

	return buf.Bytes(), err
}

func DecodeNode(raw Data) (Node, error) {
	reader := bytes.NewReader(raw)
	typ, err := readNodeType(reader)
	if err != nil {
		return nil, err
	}

	var node Node = nil
	switch typ {
	case Branch:
		node = NewBranchNode()
		err = node.Decode(reader)
	case Leaf:
		node = NewLeafNode()
		err = node.Decode(reader)
	case Extension:
		node = NewExtensionNode()
		err = node.Decode(reader)
	default:
		err = ErrInvalidNodeType
	}
	if err == nil {
		node.Details().EncodedData = raw
		node.Details().Hash, err = sha3Hash(raw)
	}
	return node, err
}

func (details *NodeDetails) Details() *NodeDetails {
	return details
}

func (details *NodeDetails) EncodeAndHash() error {
	b, err := EncodeNode(details._node)
	if err != nil {
		return err
	}
	details.EncodedData = b
	h, err := sha3Hash(b)
	if err != nil {
		return err
	}
	details.Hash = h
	return nil
}
