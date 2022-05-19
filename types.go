package main

import (
	"bytes"
	"fmt"
)

const defaultLinesCapacityPerValue = 100

type (
	// line is a representation of a log line, it contains the position and size of the line inside the buffer.
	// the objective of this structure is to avoid allocating the entire line in memory for every key-value index.
	line struct {
		startOffset int
		endOffset   int
	}

	// lines are used to embbed functions to line slices.
	lines []line

	// a field is a log key=value, or { key: value }.
	field struct {
		key   []byte
		value []byte
	}

	// a leaf is a group of lines per value.
	// Example:
	//	L1: level=error msg=hello
	//	L2: level=error msg=failed
	//	leaf: { value: "error", line: [L1, L2] }
	leaf struct {
		Value []byte
		Lines []line
	}

	leafs []leaf

	// a node is a group of keys per leafs
	// Example:
	//	L1: level=info msg=hello
	//	L2: level=error msg=failed
	//	indexer: [ level: [{ value: "info", line: [L1] }, { value: "error", line: [L2] }]]
	node struct {
		values leafs // must be sorted to perform binary search.
		len    int   // counter used to prevent recounting on len calls.
	}

	indexer interface {
		Index(fields []field, index int, size int)
		Find([]field) []line
		Stats() Stats
	}
)

func (l line) Format(s fmt.State, c rune) {
	s.Write([]byte(fmt.Sprintf("{i:%d}", l.startOffset)))
}

func (f field) Format(s fmt.State, c rune) {
	s.Write([]byte(fmt.Sprintf("{'%s':'%s'}", f.key, f.value)))
}

func (l leaf) Format(s fmt.State, c rune) {
	s.Write([]byte(fmt.Sprintf("{%v:%v}", l.Value, l.Lines)))
}

func (n *node) insertAt(index int, value []byte) *leaf {
	// we need to copy value otherwise all node entries will share the same key address.
	_value := make([]byte, len(value))
	copy(_value, value)
	leaf := leaf{
		Value: _value,
		Lines: make([]line, 0, defaultLinesCapacityPerValue),
	}
	n.values = n.values.insertAt(index, leaf, n.len)
	n.len++
	return &n.values[index]
}

func (n *node) findOrCreateLeaf(value []byte) *leaf {
	index, leaf := n.values.binarySearch(value)
	if leaf == nil {
		return n.insertAt(index, value)
	}
	return leaf
}

func (l leafs) Len() int {
	return len(l)
}

func (l leafs) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l leafs) Less(i, j int) bool {
	return bytes.Compare(l[i].Value, l[j].Value) == -1
}

// insert appends the leaf at the index position, while avoiding memory allocations.
func (a leafs) insertAt(index int, value leaf, size int) []leaf {
	if size == index { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...) // index < len(a)
	a[index] = value
	return a
}

// BinarySearch implements the iterative binary search.
// It returns a Leaf if the value matches, or the position in which that value should be inside the slice.
func (leafs leafs) binarySearch(v []byte) (int, *leaf) {
	len := len(leafs)
	if len == 0 {
		return 0, nil
	}
	if len == 1 {
		switch bytes.Compare(v, leafs[0].Value) {
		case 0:
			return 0, &leafs[0]
		case 1:
			return 1, nil
		default:
			return 0, nil
		}
	}
	low := 0
	high := len - 1
	for {
		mid := (low + high) / 2
		switch bytes.Compare(v, leafs[mid].Value) {
		case 0:
			return mid, &leafs[mid]
		case 1:
			low = mid + 1
		default:
			high = mid - 1
		}
		if high <= low {
			switch bytes.Compare(v, leafs[low].Value) {
			case 0:
				return low, &leafs[low]
			case 1:
				return low + 1, nil
			default:
				return low, nil
			}
		}
	}
}

func (lines lines) binarySearch(value int) (int, *line) {
	len := len(lines)
	if len == 0 {
		return 0, nil
	}
	if len == 1 {
		if value == lines[0].startOffset {
			return 0, &lines[0]
		}
		if value > lines[0].startOffset {
			return 1, nil
		} else {
			return 0, nil
		}
	}
	low := 0
	high := len - 1
	for {
		mid := (low + high) / 2
		if value == lines[mid].startOffset {
			return mid, &lines[mid]
		}
		if value > lines[mid].startOffset {
			low = mid + 1
		} else {
			high = mid - 1
		}
		if high <= low {
			if value == lines[low].startOffset {
				return low, &lines[low]
			}
			if value > lines[low].startOffset {
				return low + 1, nil
			} else {
				return low, nil
			}
		}
	}
}
