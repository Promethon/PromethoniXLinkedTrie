package PromethoniXTrie

type Trie interface {
	IsEmpty() bool
	Get(Hash) (Data, error)
	Put(Hash, Data) (Hash, error)
	Delete(Hash) (Hash, error)
}
