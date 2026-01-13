package datastructures

import "iter"

type TrieNode[T any] struct {
	children map[rune]*TrieNode[T]
	isLeaf   bool
	val      T
}

func NewTrieNode[T any]() TrieNode[T] {
	return TrieNode[T]{children: make(map[rune]*TrieNode[T], 27)}
}

func (node *TrieNode[T]) IsLeaf() bool {
	return node.isLeaf
}

func (node *TrieNode[T]) Val() T {
	return node.val
}

func (root *TrieNode[T]) Insert(key string, val T) {
	cur := root
	for _, char := range key {
		if cur.children[char] == nil {
			node := &TrieNode[T]{children: make(map[rune]*TrieNode[T], 26)}
			cur.children[char] = node
			cur.isLeaf = false
		}
		cur = cur.children[char]
	}
	cur.isLeaf = true
	cur.val = val
}

func (root *TrieNode[T]) SearchExact(key string) (T, bool) {
	var result T
	cur := root
	for _, char := range key {
		if cur.children[char] == nil {
			return result, false
		}
		cur = cur.children[char]
	}
	return cur.val, cur.isLeaf
}

// Returns either longest prefix match, or zero value of a type
func (root *TrieNode[T]) Search(key string) T {
	var result T
	cur := root
	for _, char := range key {
		if cur.children[char] == nil {
			return result
		}
		cur = cur.children[char]
		result = cur.val
	}
	return result
}

// Iterates over longest prefix of searched substring
func (root *TrieNode[T]) Iterate(key string) iter.Seq[*TrieNode[T]] {
	cur := root
	return func(yield func(*TrieNode[T]) bool) {
		if len(key) == 0 {
			yield(cur)
			return
		}
		for _, char := range key {
			if !yield(cur) {
				return
			}
			if cur.children[char] == nil {
				return
			}
			cur = cur.children[char]
		}
		yield(cur)
	}
}
