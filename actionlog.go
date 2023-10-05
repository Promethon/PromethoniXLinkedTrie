package PromethoniXTrie

import "fmt"

type ActionType int

const (
	Insert ActionType = iota
	Update
	Delete
)

// ActionLogEntry in log, [key, old value, new value]
type ActionLogEntry struct {
	Action  ActionType
	Key     Hash
	OldData Data
	NewData Data
}

type ActionLog struct {
	ActionLogEntries   []*ActionLogEntry
	IsActionLogEnabled bool
}

func (entry *ActionLogEntry) String() string {
	var typ string
	switch entry.Action {
	case Insert:
		typ = "Insert"
	case Update:
		typ = "Update"
	case Delete:
		typ = "Delete"
	default:
		typ = "Unknown"
	}

	return fmt.Sprintf("[Action:%s, Key:%s]", typ, string(entry.Key))
}
