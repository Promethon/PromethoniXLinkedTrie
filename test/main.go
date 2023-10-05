package main

import (
	"PromethoniXTrie"
	"fmt"
)

var blockNumber int64 = 1

func main() {
	t, _ := PromethoniXTrie.NewLinkedTrieImpl(
		true,
		2600000,
		func() int64 {
			return blockNumber
		},
	)
	fmt.Println("LevelDB:")
	debug(t)

	fmt.Println("\nTest:")

	blockNumber = 5
	t.Put(PromethoniXTrie.Hash("Test1"), PromethoniXTrie.Data("Data1"))
	debug(t)

	blockNumber = 13000
	t.Put(PromethoniXTrie.Hash("Test2"), PromethoniXTrie.Data("Data2"))
	debug(t)

	blockNumber = 1700000
	t.Put(PromethoniXTrie.Hash("Test3"), PromethoniXTrie.Data("Data3"))
	debug(t)

	blockNumber = 1800000
	t.Put(PromethoniXTrie.Hash("Test1"), PromethoniXTrie.Data("Data1New"))
	debug(t)

	blockNumber = 2900000
	fmt.Println("\nRemoving:")
	t.RemoveOldNodes()
	debug(t)

	fmt.Println("\nActionLogEntries:")
	fmt.Println(t.ActionLogEntries())
}

func debug(trie *PromethoniXTrie.LinkedTrieImpl) {
	i := trie.Iterator()
	var err error = nil
	var data PromethoniXTrie.Data = nil

	for err == nil {
		data, err = i()
		if err == nil {
			fmt.Print(string(data) + " -> ")
		}
	}
	fmt.Println()
}
