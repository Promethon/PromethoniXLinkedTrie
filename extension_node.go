package PromethoniXTrie

import "io"

type ExtensionNode struct {
	NodeDetails
	Path     Hash
	NextHash Hash
}

func (node *ExtensionNode) Type() NodeType {
	return Extension
}

func (node *ExtensionNode) Encode(writer io.Writer) error {
	var err error

	l := len(node.Path)
	err = writeInt32(writer, int32(l))
	if err == nil && l > 0 {
		_, err = writer.Write(node.Path)
	}
	if err == nil {
		l = len(node.NextHash)
		err = writeInt32(writer, int32(l))
	}
	if err == nil && l > 0 {
		_, err = writer.Write(node.NextHash)
	}
	return err
}

func (node *ExtensionNode) Decode(reader io.Reader) error {
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
	node.NextHash, err = readBytes(reader, length)
	return nil
}

func (node *ExtensionNode) NextRoute(route Route) (Hash, Route, error) {
	matchLen := prefixLen(node.Path, route)
	if matchLen != len(node.Path) {
		return node.NextHash, nil, ErrNotFound
	}
	return node.NextHash, route[matchLen:], nil
}

func AsExtension(node Node) *ExtensionNode {
	return node.(*ExtensionNode)
}

func NewExtensionNode() *ExtensionNode {
	node := new(ExtensionNode)
	node._node = node
	return node
}
