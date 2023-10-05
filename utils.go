package PromethoniXTrie

import (
	"encoding/binary"
	"golang.org/x/crypto/sha3"
	"io"
	"unsafe"
)

func writeNodeType(writer io.Writer, value NodeType) error {
	return binary.Write(writer, binary.LittleEndian, value)
}

func readNodeType(reader io.Reader) (NodeType, error) {
	var value NodeType
	err := binary.Read(reader, binary.LittleEndian, &value)
	return value, err
}

func writeInt32(writer io.Writer, value int32) error {
	return binary.Write(writer, binary.LittleEndian, value)
}

func readInt32(reader io.Reader) (int32, error) {
	var value int32
	err := binary.Read(reader, binary.LittleEndian, &value)
	return value, err
}

func readBytes(reader io.Reader, len int32) ([]byte, error) {
	if len == 0 {
		return nil, nil
	}

	bytes := make([]byte, len)
	_, err := reader.Read(bytes)
	return bytes, err
}

// sha3Hash Calculates the SHA3-256 hash of the input data.
func sha3Hash(input Data) (Hash, error) {
	hash := sha3.New256()
	_, err := hash.Write(input)
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

// keyToRoute returns hex bytes
// e.g {0xa1, 0xf2} -> {0xa, 0x1, 0xf, 0x2}
func keyToRoute(key Hash) Route {
	l := len(key) * 2
	var route = make(Route, l)
	for i, b := range key {
		route[i*2] = b / 16
		route[i*2+1] = b % 16
	}
	return route
}

// routeToKey returns native bytes
// e.g {0xa, 0x1, 0xf, 0x2} -> {0xa1, 0xf2}
func routeToKey(route Route) Hash {
	l := len(route) / 2
	var key = make(Hash, l)
	for i := 0; i < l; i++ {
		key[i] = route[i*2]<<4 + route[i*2+1]
	}
	return key
}

// prefixLen returns the length of the common prefix between a and b.
func prefixLen(a, b []byte) int {
	var i, length = 0, min(len(a), len(b))
	for ; i < length; i++ {
		if a[i] != b[i] {
			break
		}
	}
	return i
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func ptrToByteSlice(u uintptr) []byte {
	size := unsafe.Sizeof(u)
	b := make([]byte, size)
	switch size {
	case 4:
		binary.LittleEndian.PutUint32(b, uint32(u))
	case 8:
		binary.LittleEndian.PutUint64(b, uint64(u))
	}
	return b
}

func byteSliceToPtr(b []byte) uintptr {
	switch len(b) {
	case 4:
		return uintptr(binary.LittleEndian.Uint32(b))
	case 8:
		return uintptr(binary.LittleEndian.Uint64(b))
	}
	return 0
}
