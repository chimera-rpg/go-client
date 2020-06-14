package data

import (
	"hash/crc32"
)

var stringMapTable = crc32.MakeTable(crc32.Koopman)

// StringID is a unique ID for a particular string
type StringID = uint32

// StringsMap provides a StringID to string map and reverse map.
type StringsMap struct {
	IDs     map[StringID]string
	Strings map[string]StringID
}

// Acquire returns the StringID that the provided name string corresponds to.
func (n *StringsMap) Acquire(name string) StringID {
	if val, ok := n.Strings[name]; ok {
		return val
	}
	id := crc32.Checksum([]byte(name), stringMapTable)

	n.IDs[id] = name
	n.Strings[name] = id

	return id
}

// Lookup reutrns the string that the provided StringID corresponds to.
func (n *StringsMap) Lookup(id StringID) string {
	if val, ok := n.IDs[id]; ok {
		return val
	}
	return ""
}

// NewStringsMap provides a constructed instance of StringsMap.
func NewStringsMap() StringsMap {
	return StringsMap{
		IDs:     make(map[StringID]string),
		Strings: make(map[string]StringID),
	}
}
