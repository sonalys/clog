package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
)

// Default capacity for number of different values in each key.
const defaultValuesCapacity = 10000

type mapIndexer struct {
	// Maps fields -> values -> lines
	// Since fields are not frequently inserted, we use hash map for o(1) read so we get to the field value index faster.
	// After finding the key index, we perform binary search to find which value we append the lines to.
	keys map[string]node
}

func (i *mapIndexer) Index(fields []field, index int, size int) {
	for _, field := range fields {
		var keyIndex node
		key := string(field.key)
		if val, ok := i.keys[key]; !ok {
			keyIndex = node{
				values: make([]leaf, 0, defaultValuesCapacity),
			}
		} else {
			keyIndex = val
		}
		leaf := keyIndex.findOrCreateLeaf(field.value)
		leaf.Lines = append(leaf.Lines, line{startOffset: index, endOffset: index + size})
		i.keys[key] = keyIndex
	}
}

// Find returns the lines that contains all the fields and respective values.
func (indexer *mapIndexer) Find(fields []field) (resp []line) {
	findings := make([]lines, len(fields))
	// find all individual field matching lines.
	for i := range fields {
		f := &fields[i]
		if index, ok := indexer.keys[string(f.key)]; !ok {
			return
		} else {
			findings[i] = append(findings[i], index.findOrCreateLeaf(f.value).Lines...)
		}
	}

	if len(fields) == 1 {
		return findings[0]
	}

	// check that the lines are matched by all fields.
	for i := 0; i < len(fields)-1; i++ {
		for j := range findings[i] {
			found := false
			for k := i + 1; k < len(fields); k++ {
				foundIndex, _ := findings[k].binarySearch(findings[i][j].startOffset)
				if foundIndex > -1 { // if found, mark as found and remove from future iterations.
					findings[k] = append(findings[k][:foundIndex], findings[k][foundIndex+1:]...)
					found = true
				}
			}
			if found {
				resp = append(resp, findings[i][j])
			}
		}
	}
	return
}

type NodeStats struct {
	name       string
	lineCount  int
	valueCount int
}

func (f NodeStats) Format(s fmt.State, c rune) {
	s.Write([]byte(fmt.Sprintf(
		"%s\n{Lines:%d Values:%d}\n",
		color.RedString(f.name),
		f.lineCount,
		f.valueCount,
	)))
}

type Stats struct {
	nodes     []NodeStats
	nodeCount int
}

func (f Stats) Format(s fmt.State, c rune) {
	s.Write([]byte(fmt.Sprintf("Keys: %d\n\n", f.nodeCount)))
	for _, v := range f.nodes {
		s.Write([]byte(fmt.Sprintf("%s", v)))
	}
}

func (indexer *mapIndexer) Stats() (stats Stats) {
	stats.nodes = make([]NodeStats, 0, len(indexer.keys))
	for k, v := range indexer.keys {
		stats.nodeCount++
		lineCount := 0
		for i := range v.values {
			v := &v.values[i]
			lineCount += len(v.Lines)
		}
		stats.nodes = append(stats.nodes, NodeStats{
			name:       k,
			lineCount:  lineCount,
			valueCount: v.len,
		})
	}

	sort.Slice(stats.nodes, func(i, j int) bool {
		diff := stats.nodes[i].lineCount - stats.nodes[j].lineCount
		switch {
		case diff < 0:
			return false
		case diff == 0:
			return strings.Compare(stats.nodes[i].name, stats.nodes[j].name) < 0
		default:
			return true
		}
	})
	return
}
