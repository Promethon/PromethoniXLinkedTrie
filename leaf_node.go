package PromethoniXTrie

import "io"

type LeafNode struct {
	NodeDetails
	Path  Hash
	Value Data
}

func (node *LeafNode) Type() NodeType {
	return Leaf
}

func (node *LeafNode) Encode(writer io.Writer) error {
	var err error

	l := len(node.Path)
	err = writeInt32(writer, int32(l))
	if err == nil && l > 0 {
		_, err = writer.Write(node.Path)
	}
	if err == nil {
		l = len(node.Value)
		err = writeInt32(writer, int32(l))
	}
	if err == nil && l > 0 {
		_, err = writer.Write(node.Value)
	}
	return err
}

func (node *LeafNode) Decode(reader io.Reader) error {
	length, err := readInt32(reader)
	if err != nil {
		return err
	}
	node.Path, err = readBytes(reader, length)
	if err != nil {
		return err
	}
	length, err = readInt32(reader)
	if err != nil {
		return err
	}
	node.Value, err = readBytes(reader, length)
	return nil
}

func (node *LeafNode) NextRoute(Route) (Hash, Route, error) {
	return nil, nil, nil
}

func AsLeaf(node Node) *LeafNode {
	return node.(*LeafNode)
}

func NewLeafNode() *LeafNode {
	node := new(LeafNode)
	node._node = node
	return node
}
