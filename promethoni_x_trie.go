package PromethoniXTrie

import (
	"bytes"
	"github.com/nacamp/go-simplechain/storage"
)

// PromethoniXTrie is a Merkle Patricia Trie, consists of three kinds of nodes,
// Branch Node: 16-elements array, value is [hash_0, hash_1, ..., hash_f, hash]
// Extension Node: 3-elements array, value is [ext flag, prefix path, next hash]
// Leaf Node: 3-elements array, value is [leaf flag, suffix path, value]
type PromethoniXTrie struct {
	ActionLog
	rootHash Hash
	storage  storage.Storage
}

var rootHashKey = []byte{0, 0, 1, 1}

func NewPromethoniXTrie(
	isActionLogEnabled bool,
) (*PromethoniXTrie, error) {
	dbStorage, err := storage.NewLevelDBStorage("./db")
	if err != nil {
		return nil, err
	}

	rootHash, err := dbStorage.Get(rootHashKey)
	if err != nil && err != ErrNotFound {
		return nil, err
	}

	t := &PromethoniXTrie{
		rootHash: rootHash,
		storage:  dbStorage,
		ActionLog: ActionLog{
			IsActionLogEnabled: isActionLogEnabled,
		},
	}
	if t.rootHash == nil || len(t.rootHash) == 0 {
		return t, nil
	} else if _, err := t.storage.Get(rootHash); err != nil {
		return nil, err
	}
	return t, nil
}

func (trie *PromethoniXTrie) IsEmpty() bool {
	return trie.rootHash == nil
}

func (trie *PromethoniXTrie) RootHash() Hash {
	return trie.rootHash
}

// CommitNode node in trie into storage
func (trie *PromethoniXTrie) commitNode(n Node) error {
	if err := n.EncodeAndHash(); err != nil {
		return err
	}
	return trie.storage.Put(n.Details().Hash, n.Details().EncodedData)
}

func (trie *PromethoniXTrie) fetchNode(hash Hash) (Node, error) {
	raw, err := trie.storage.Get(hash)
	if err != nil {
		return nil, err
	}

	return DecodeNode(raw)
}

func (trie *PromethoniXTrie) Get(key Hash) (Data, error) {
	rootHash := trie.rootHash
	route := keyToRoute(key)

	for len(route) >= 0 {
		rootNode, err := trie.fetchNode(rootHash)
		if err != nil {
			return nil, err
		}
		if rootNode.Type() == Leaf {
			if !bytes.Equal(AsLeaf(rootNode).Path, route) {
				break
			}
			return AsLeaf(rootNode).Value, nil
		} else if len(route) == 0 {
			break
		} else {
			rootHash, route, err = rootNode.NextRoute(route)
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, ErrNotFound
}

func (trie *PromethoniXTrie) Put(key Hash, value Data) (Hash, error) {
	var oldData Data = nil
	var err error
	action := Update

	if trie.IsActionLogEnabled {
		oldData, err = trie.Get(key)
		if err != nil {
			action = Insert
		}
	}

	newHash, err := trie.update(trie.rootHash, keyToRoute(key), value)
	if err != nil {
		return nil, err
	}
	trie.rootHash = newHash

	err = trie.storage.Put(rootHashKey, newHash)
	if err != nil {
		return nil, err
	}

	if trie.IsActionLogEnabled {
		entry := &ActionLogEntry{action, key, oldData, value}
		trie.ActionLogEntries = append(trie.ActionLogEntries, entry)
	}

	return newHash, nil
}

// Delete the node's value in trie
/*
	1. ext(ext->leaf-->leaf,ext->ext--->ext)
	2. branch(branch->leaf-->leaf,branch->branch-->ext->branch,branch->ext-->ext)

	 ext		 ext
	  | 		  |
	branch 	-->	 leaf	-->	leaf
	/	\
[leaf]	leaf

  	branch					 ext
	/	\					  |
[leaf]	ext		--> 	    branch
		 |
	   branch

  	branch					 ext
	/	\					  |
[leaf]	branch		--> 	branch
		/	\				/	\
		leaf leaf			leaf leaf

*/
func (trie *PromethoniXTrie) Delete(key Hash) (Hash, error) {
	var oldData Data = nil
	if trie.IsActionLogEnabled {
		oldData, _ = trie.Get(key)
	}

	newHash, err := trie.delete(trie.rootHash, keyToRoute(key))
	if err != nil {
		return nil, err
	}
	trie.rootHash = newHash

	err = trie.storage.Put(rootHashKey, newHash)
	if err != nil {
		return nil, err
	}

	if trie.IsActionLogEnabled {
		entry := &ActionLogEntry{Delete, key, oldData, nil}
		trie.ActionLogEntries = append(trie.ActionLogEntries, entry)
	}
	return newHash, nil
}
