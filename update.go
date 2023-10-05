package PromethoniXTrie

func (trie *PromethoniXTrie) update(root Hash, route Route, value Data) (Hash, error) {
	if len(root) == 0 {
		// directly add leaf node
		node := NewLeafNode()
		node.Path = Hash(route)
		node.Value = value
		err := trie.commitNode(node)
		if err != nil {
			return nil, err
		}
		return node.Hash, nil
	}

	rootNode, err := trie.fetchNode(root)
	if err != nil {
		return nil, err
	}

	switch node := rootNode.(type) {
	case *BranchNode:
		return trie.updateBranch(node, route, value)
	case *ExtensionNode:
		return trie.updateExtension(node, route, value)
	case *LeafNode:
		return trie.updateLeaf(node, route, value)
	}
	return nil, ErrUnexpected
}

// add new node to one branch of branch node's 16 branches according to route
func (trie *PromethoniXTrie) updateBranch(node *BranchNode, route Route, value Data) (Hash, error) {
	nextHash, nextRoute, err := node.NextRoute(route)
	if err != nil {
		return nil, err
	}

	// update sub-trie
	newHash, err := trie.update(nextHash, nextRoute, value)
	if err != nil {
		return nil, err
	}

	// update the branch hash
	node.Hashes[route[0]] = newHash

	// save updated node to storage
	err = trie.commitNode(node)
	return node.Hash, err
}

func (trie *PromethoniXTrie) updateExtension(node *ExtensionNode, route Route, value Data) (Hash, error) {
	path := node.Path
	if len(path) > len(route) {
		return nil, ErrWrongKey
	}

	// add new node to the ext node's sub-trie
	nextHash, nextRoute, err := node.NextRoute(route)
	if err == nil {
		newHash, err := trie.update(nextHash, nextRoute, value)
		if err != nil {
			return nil, err
		}

		// update the new hash
		node.NextHash = newHash

		// save updated node to storage
		err = trie.commitNode(node)
		return node.Hash, err
	}

	brNode := NewBranchNode()
	if err = brNode.EncodeAndHash(); err != nil {
		return nil, err
	}

	matchLen := prefixLen(path, route)
	if matchLen+1 < len(path) {
		extNode := NewExtensionNode()
		extNode.Path = path[matchLen+1:]
		extNode.NextHash = nextHash

		if err = trie.commitNode(extNode); err != nil {
			return nil, err
		}
		brNode.Hashes[path[matchLen]] = extNode.Hash
	} else {
		brNode.Hashes[path[matchLen]] = nextHash
	}

	// a branch to hold the new node
	brNode.Hashes[route[matchLen]], err = trie.update(nil, route[matchLen+1:], value)
	if err != nil {
		return nil, err
	}

	// save branch to the storage
	if err = trie.commitNode(brNode); err != nil {
		return nil, err
	}

	// if no common prefix, replace the ext node with the new branch node
	if matchLen == 0 {
		return brNode.Hash, nil
	}

	// use the new branch node as the ext node's sub-trie
	node.Path = path[0:matchLen]
	node.NextHash = brNode.Hash

	if err = trie.commitNode(node); err != nil {
		return nil, err
	}
	return node.Hash, nil
}

// split leaf node's into an ext node and a branch node based on
// the longest common prefix between route and leaf node's path
// add new node to the branch node
func (trie *PromethoniXTrie) updateLeaf(node *LeafNode, route Route, value Data) (Hash, error) {
	var err error
	path := node.Path
	leafVal := node.Value
	if len(path) > len(route) {
		return nil, ErrWrongKey
	}

	matchLen := prefixLen(path, route)
	// node exists, update its value
	if matchLen == len(path) {

		if len(route) > matchLen {
			return nil, ErrWrongKey
		}

		node.Value = value
		// save updated node to storage
		err = trie.commitNode(node)
		return node.Hash, err
	}

	// create a new branch for the new node
	brNode := NewBranchNode()
	if err = brNode.EncodeAndHash(); err != nil {
		return nil, err
	}

	// a branch to hold the leaf node
	brNode.Hashes[path[matchLen]], err = trie.update(nil, Route(path[matchLen+1:]), leafVal)
	if err != nil {
		return nil, err
	}

	// a branch to hold the new node
	brNode.Hashes[route[matchLen]], err = trie.update(nil, route[matchLen+1:], value)
	if err != nil {
		return nil, err
	}

	// save the new branch node to storage
	if err := trie.commitNode(brNode); err != nil {
		return nil, err
	}

	// if no common prefix, replace the leaf node with the new branch node
	if matchLen == 0 {
		return brNode.Hash, nil
	}

	// create a new ext node, and use the new branch node as the new ext node's sub-trie
	extNode := NewExtensionNode()
	extNode.Path = path[0:matchLen]
	extNode.NextHash = brNode.Hash

	if err := trie.commitNode(extNode); err != nil {
		return nil, err
	}
	return extNode.Hash, nil
}
