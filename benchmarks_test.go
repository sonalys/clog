package main

import (
	"fmt"
	"testing"
)

func Benchmark_node_insertValue(b *testing.B) {
	f := func(b *testing.B, size int) {
		n := node{
			values: make([]leaf, size, size+1),
			len:    size,
		}
		for i := range n.values {
			n.values[i].Value = []byte{byte(i)}
		}
		b.ResetTimer()
		n.insertAt(size/2, []byte{byte(size / 2)})
	}
	for i := 100; i <= 1_000_000; i *= 10 {
		b.Run(fmt.Sprintf("%d_possible_values", i), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				f(b, i)
			}
		})
	}
}

func Benchmark_node_findOrCreateLeaf(b *testing.B) {
	f := func(b *testing.B, size int) {
		n := node{
			values: make([]leaf, size),
			len:    size,
		}
		for i := range n.values {
			n.values[i] = leaf{Value: []byte{byte(i)}}
		}
		key := n.values[size-1].Value
		b.ResetTimer()
		n.findOrCreateLeaf(key)
	}
	b.StopTimer()
	for i := 100; i <= 1_000_000; i *= 10 {
		b.Run(fmt.Sprintf("%d_possible_values", i), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				f(b, i)
			}
		})
	}
}

func Benchmark_leaf_binarySearch(b *testing.B) {
	f := func(b *testing.B, size int) {
		slice := make(leafs, size)
		for i := 0; i < size; i++ {
			slice[i] = leaf{
				Value: []byte{byte(i)},
			}
		}
		b.ResetTimer()
		slice.binarySearch([]byte{byte(size - 1)})
	}
	for i := 100; i <= 1_000_000; i *= 10 {
		b.Run(fmt.Sprintf("%d_leafs", i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				f(b, i)
			}
		})
	}
}
