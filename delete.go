package PromethoniXTrie

func (trie *PromethoniXTrie) delete(root Hash, route Route) (Hash, error) {
	if len(root) == 0 {
		return nil, ErrNotFound
	}

	// fetch sub-trie root node
	rootNode, err := trie.fetchNode(root)
	if err != nil {
		return nil, err
	}

	switch node := rootNode.(type) {
	case *BranchNode:
		return trie.deleteBranch(node, route)
	case *ExtensionNode:
		return trie.deleteExtension(node, route)
	case *LeafNode:
		return trie.deleteLeaf(node, route)
	default:
		return nil, ErrUnexpected
	}
}

func (trie *PromethoniXTrie) deleteBranch(node *BranchNode, route Route) (Hash, error) {
	nextHash, nextRoute, err := node.NextRoute(route)
	if err != nil {
		return nil, err
	}

	newHash, err := trie.delete(nextHash, nextRoute)
	if err != nil {
		return nil, err
	}

	node.Hashes[route[0]] = newHash

	// remove empty branch node
	branchLength := node.Length()
	if branchLength == 0 {
		return nil, nil
	}
	if branchLength == 1 {
		return trie.deleteSingleBranch(node)
	}

	err = trie.commitNode(node)
	return node.Hash, err
}

func (trie *PromethoniXTrie) deleteSingleBranch(rootNode *BranchNode) (Hash, error) {
	for idx := range rootNode.Hashes {
		if len(rootNode.Hashes[idx]) != 0 {

			childNode, err := trie.fetchNode(rootNode.Hashes[idx])
			if err != nil {
				return nil, err
			}

			switch child := childNode.(type) {
			case *BranchNode:
				extNode := NewExtensionNode()
				extNode.Path = Hash{byte(idx)}
				extNode.NextHash = rootNode.Hashes[idx]

				err = trie.commitNode(extNode)
				return extNode.Hash, err

			case *ExtensionNode:
				child.Path = append(Hash{byte(idx)}, child.Path...)
				err = trie.commitNode(child)
				return child.Hash, err

			case *LeafNode: // branch->leaf-->leaf
				child.Path = append(Hash{byte(idx)}, child.Path...)
				err = trie.commitNode(child)
				return child.Hash, err

			default:
				return nil, ErrUnexpected
			}
		}
	}
	return nil, nil
}

func (trie *PromethoniXTrie) deleteExtension(node *ExtensionNode, route Route) (Hash, error) {
	matchLen := prefixLen(node.Path, route)
	if matchLen != len(node.Path) {
		return nil, ErrNotFound
	}

	childHash, err := trie.delete(node.NextHash, route[matchLen:])
	if err != nil {
		return nil, err
	}

	// remove empty ext node
	if childHash == nil {
		return nil, nil
	}

	// child hash
	var newHash Hash
	newHash, err = trie.deleteSingleExtension(node, childHash)
	if err != nil {
		return nil, err
	}
	if newHash != nil {
		return newHash, nil
	}

	node.NextHash = childHash
	err = trie.commitNode(node)
	return node.Hash, err
}

func (trie *PromethoniXTrie) deleteSingleExtension(node *ExtensionNode, hash Hash) (Hash, error) {
	childNode, err := trie.fetchNode(hash)
	if err != nil {
		return nil, err
	}

	switch child := childNode.(type) {
	case *ExtensionNode:
		child.Path = append(node.Path, child.Path...)
		err = trie.commitNode(child)
		return child.Hash, err

	case *LeafNode:
		child.Path = append(node.Path, child.Path...)
		err = trie.commitNode(child)
		return child.Hash, err

	default:
		return nil, nil
	}
}

func (trie *PromethoniXTrie) deleteLeaf(node *LeafNode, route Route) (Hash, error) {
	matchLen := prefixLen(node.Path, route)
	if matchLen != len(node.Path) {
		return nil, ErrNotFound
	}
	return nil, nil
}
