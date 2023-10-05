package PromethoniXTrie

import (
	"bytes"
	"encoding/gob"
)

var storageOldestNodeKey = []byte{0, 1, 2}
var storageNewestNodeKey = []byte{0, 2, 4}

type linkedNode struct {
	NextNode     Hash
	PrevNode     Hash
	Data         Data
	LastModified int64
}

type LinkedTrieImpl struct {
	trie               *PromethoniXTrie
	blockNumberFunc    func() int64
	maxDiffBlockNumber int64
	oldestNodeKey      Hash
	newestNodeKey      Hash
}

func NewLinkedTrieImpl(
	isActionLogEnabled bool,
	maxDiffBlockNumber int64,
	blockNumberFunc func() int64,
) (*LinkedTrieImpl, error) {
	trie, err := NewPromethoniXTrie(isActionLogEnabled)
	if err != nil {
		return nil, err
	}

	oldestNodeKey, err := trie.storage.Get(storageOldestNodeKey)
	if err != nil {
		oldestNodeKey = nil
	}

	newestNodeKey, err := trie.storage.Get(storageNewestNodeKey)
	if err != nil {
		newestNodeKey = nil
	}

	trieTree := &LinkedTrieImpl{
		trie:               trie,
		blockNumberFunc:    blockNumberFunc,
		maxDiffBlockNumber: maxDiffBlockNumber,
		oldestNodeKey:      oldestNodeKey,
		newestNodeKey:      newestNodeKey,
	}
	return trieTree, nil
}

func (node *linkedNode) encode() (Data, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(node)
	return buf.Bytes(), err
}

func (data Data) decode() (*linkedNode, error) {
	dec := gob.NewDecoder(bytes.NewReader(data))
	node := linkedNode{}
	err := dec.Decode(&node)
	return &node, err
}

func (linkedTrie *LinkedTrieImpl) IsEmpty() bool {
	return linkedTrie.trie.IsEmpty()
}

func (linkedTrie *LinkedTrieImpl) getNode(key Hash) (*linkedNode, error) {
	data, err := linkedTrie.trie.Get(key)
	if err != nil {
		return nil, err
	}

	return data.decode()
}

func (linkedTrie *LinkedTrieImpl) Get(key Hash) (Data, error) {
	node, err := linkedTrie.getNode(key)
	if err != nil {
		return nil, err
	}

	return node.Data, nil
}

func (linkedTrie *LinkedTrieImpl) putNode(key Hash, node *linkedNode) (Hash, error) {
	raw, err := node.encode()
	if err != nil {
		return nil, err
	}
	return linkedTrie.trie.Put(key, raw)
}

func (linkedTrie *LinkedTrieImpl) Put(key Hash, data Data) (Hash, error) {
	node, err := linkedTrie.getNode(key)
	if err == ErrNotFound {
		node = &linkedNode{}
	} else if node == nil {
		return nil, err
	} else {
		if node.NextNode == nil {
			node.NextNode = linkedTrie.newestNodeKey
		}
		err = linkedTrie.update(node)
		if err != nil {
			return nil, err
		}
	}
	nodeNextKey := node.NextNode
	node.LastModified = linkedTrie.blockNumberFunc()
	node.NextNode = nil
	if !bytes.Equal(key, linkedTrie.newestNodeKey) {
		node.PrevNode = linkedTrie.newestNodeKey
	}
	node.Data = data

	newHash, err := linkedTrie.putNode(key, node)
	if err == nil {
		if linkedTrie.newestNodeKey == nil {
			linkedTrie.newestNodeKey = key
			err = linkedTrie.trie.storage.Put(storageNewestNodeKey, key)
			if err != nil {
				return newHash, err
			}
		} else if linkedTrie.newestNodeKey != nil && !bytes.Equal(linkedTrie.newestNodeKey, key) {
			newestNode, err := linkedTrie.getNode(linkedTrie.newestNodeKey)
			if err != nil {
				return newHash, err
			}
			newestNode.NextNode = key
			_, err = linkedTrie.putNode(linkedTrie.newestNodeKey, newestNode)
			if err != nil {
				return newHash, err
			}
		}

		if linkedTrie.oldestNodeKey == nil {
			linkedTrie.oldestNodeKey = key
			err = linkedTrie.trie.storage.Put(storageOldestNodeKey, key)
			if err != nil {
				return newHash, err
			}

		} else if bytes.Equal(linkedTrie.oldestNodeKey, key) {
			nextOldestNode := nodeNextKey
			if nextOldestNode != nil {
				linkedTrie.oldestNodeKey = nextOldestNode
				err = linkedTrie.trie.storage.Put(storageOldestNodeKey, nextOldestNode)
				if err != nil {
					return newHash, err
				}
			}
		}

		linkedTrie.newestNodeKey = key
		err = linkedTrie.trie.storage.Put(storageNewestNodeKey, key)
	}
	return newHash, err
}

func (linkedTrie *LinkedTrieImpl) Delete(key Hash) (Hash, error) {
	if key == nil {
		return nil, ErrWrongKey
	}

	node, err := linkedTrie.getNode(key)
	if err != nil {
		return nil, err
	}
	err = linkedTrie.update(node)
	if err != nil {
		return nil, err
	}

	if bytes.Equal(linkedTrie.newestNodeKey, key) {
		linkedTrie.newestNodeKey = node.PrevNode
		err = linkedTrie.trie.storage.Put(storageNewestNodeKey, node.PrevNode)
		if err != nil {
			return nil, err
		}
	}

	if bytes.Equal(linkedTrie.oldestNodeKey, key) {
		linkedTrie.oldestNodeKey = node.NextNode
		err = linkedTrie.trie.storage.Put(storageOldestNodeKey, node.NextNode)
		if err != nil {
			return nil, err
		}
	}

	return linkedTrie.trie.Delete(key)
}

func (linkedTrie *LinkedTrieImpl) update(node *linkedNode) error {
	if node.PrevNode != nil {
		prevNode, err := linkedTrie.getNode(node.PrevNode)
		if err != nil {
			return err
		}
		prevNode.NextNode = node.NextNode
		_, err = linkedTrie.putNode(node.PrevNode, prevNode)
		if err != nil {
			return err
		}
	}

	if node.NextNode != nil {
		nextNode, err := linkedTrie.getNode(node.NextNode)
		if err != nil {
			return err
		}
		nextNode.PrevNode = node.PrevNode
		_, err = linkedTrie.putNode(node.NextNode, nextNode)
		if err != nil {
			return err
		}
	}
	return nil
}

func (linkedTrie *LinkedTrieImpl) Iterator() func() (Data, error) {
	root := linkedTrie.oldestNodeKey
	return func() (Data, error) {
		node, err := linkedTrie.getNode(root)
		if err != nil {
			return nil, err
		}
		root = node.NextNode
		return node.Data, nil
	}
}

func (linkedTrie *LinkedTrieImpl) ActionLogEntries() []*ActionLogEntry {
	return linkedTrie.trie.ActionLogEntries
}

func (linkedTrie *LinkedTrieImpl) RemoveOldNodes() error {
	for linkedTrie.oldestNodeKey != nil {
		node, err := linkedTrie.getNode(linkedTrie.oldestNodeKey)
		if err != nil {
			return err
		}

		if linkedTrie.blockNumberFunc()-node.LastModified > linkedTrie.maxDiffBlockNumber {
			_, err := linkedTrie.Delete(linkedTrie.oldestNodeKey)
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
	return nil
}
