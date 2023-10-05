package PromethoniXTrie

import (
	"errors"
	"github.com/nacamp/go-simplechain/storage"
)

var (
	ErrNotFound        = storage.ErrKeyNotFound
	ErrInvalidNodeType = errors.New("raw data is invalid")
	ErrWrongKey        = errors.New("wrong key, too short")
	ErrUnexpected      = errors.New("unexpected error")
)
