package PromethoniXTrie

import (
	"io"
)

const branchHashSize = 16

type BranchNode struct {
	NodeDetails
	Hashes [branchHashSize]Hash
}

func (node *BranchNode) Type() NodeType {
	return Branch
}

func (node *BranchNode) Encode(writer io.Writer) error {
	var err error = nil
	for i := 0; i < branchHashSize; i++ {
		l := len(node.Hashes[i])
		err = writeInt32(writer, int32(l))
		if err != nil {
			return err
		}
		if l > 0 {
			_, err = writer.Write(node.Hashes[i])
		}
		if err != nil {
			return err
		}
	}

	return err
}

func (node *BranchNode) Decode(reader io.Reader) error {
	for i := 0; i < branchHashSize; i++ {
		length, err := readInt32(reader)
		if err != nil {
			return err
		}
		node.Hashes[i], err = readBytes(reader, length)
		if err != nil {
			return err
		}
	}

	return nil
}

func (node *BranchNode) NextRoute(route Route) (Hash, Route, error) {
	return node.Hashes[route[0]], route[1:], nil
}

func (node *BranchNode) Length() int {
	l := 0
	for idx := range node.Hashes {
		if len(node.Hashes[idx]) != 0 {
			l++
		}
	}
	return l
}

func AsBranch(node Node) *BranchNode {
	return node.(*BranchNode)
}

func NewBranchNode() *BranchNode {
	node := new(BranchNode)
	node._node = node
	return node
}
