package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_lines_BinarySearch(t *testing.T) {
	tests := []struct {
		name     string
		lines    lines
		value    int
		expIndex int
		expLine  *line
	}{
		{
			name:     "empty",
			lines:    lines{},
			value:    1,
			expIndex: 0,
			expLine:  nil,
		},
		{
			name: "one, not found",
			lines: lines{
				{startOffset: 2},
			},
			value:    1,
			expIndex: 0,
			expLine:  nil,
		},
		{
			name: "one, found",
			lines: lines{
				{startOffset: 1},
			},
			value:    1,
			expIndex: 0,
			expLine:  &line{startOffset: 1},
		},
		{
			name: "first",
			lines: lines{
				{startOffset: 1},
				{startOffset: 2},
				{startOffset: 3},
			},
			value:    1,
			expIndex: 0,
			expLine:  &line{startOffset: 1},
		},
		{
			name: "middle",
			lines: lines{
				{startOffset: 1},
				{startOffset: 2},
				{startOffset: 3},
			},
			value:    2,
			expIndex: 1,
			expLine:  &line{startOffset: 2},
		},
		{
			name: "last",
			lines: lines{
				{startOffset: 1},
				{startOffset: 2},
				{startOffset: 3},
			},
			value:    3,
			expIndex: 2,
			expLine:  &line{startOffset: 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIndex, gotLine := tt.lines.binarySearch(tt.value)
			require.Equal(t, tt.expIndex, gotIndex)
			require.Equal(t, tt.expLine, gotLine)
		})
	}
}

func Test_leafs_binarySearch(t *testing.T) {
	tests := []struct {
		name     string
		leafs    leafs
		value    []byte
		expIndex int
		expLeaf  *leaf
	}{
		{
			name:     "empty",
			leafs:    leafs{},
			value:    []byte{},
			expIndex: 0,
			expLeaf:  nil,
		},
		{
			name: "one smaller, not found",
			leafs: leafs{
				{Value: []byte{0}},
			},
			value:    []byte{1},
			expIndex: 1,
			expLeaf:  nil,
		},
		{
			name: "one bigger, not found",
			leafs: leafs{
				{Value: []byte{1}},
			},
			value:    []byte{0},
			expIndex: 0,
			expLeaf:  nil,
		},
		{
			name: "one, found",
			leafs: leafs{
				{Value: []byte{1}},
			},
			value:    []byte{1},
			expIndex: 0,
			expLeaf:  &leaf{Value: []byte{1}},
		},
		{
			name: "first",
			leafs: leafs{
				{Value: []byte{1}},
				{Value: []byte{2}},
				{Value: []byte{3}},
			},
			value:    []byte{1},
			expIndex: 0,
			expLeaf:  &leaf{Value: []byte{1}},
		},
		{
			name: "middle",
			leafs: leafs{
				{Value: []byte{1}},
				{Value: []byte{2}},
				{Value: []byte{3}},
			},
			value:    []byte{2},
			expIndex: 1,
			expLeaf:  &leaf{Value: []byte{2}},
		},
		{
			name: "last",
			leafs: leafs{
				{Value: []byte{1}},
				{Value: []byte{2}},
				{Value: []byte{3}},
			},
			value:    []byte{3},
			expIndex: 2,
			expLeaf:  &leaf{Value: []byte{3}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIndex, gotLeaf := tt.leafs.binarySearch(tt.value)
			require.Equal(t, tt.expIndex, gotIndex)
			require.Equal(t, tt.expLeaf, gotLeaf)
		})
	}
}
