package lru

import (
	"testing"
)

func TestLRU(t *testing.T) {
	lru := NewLRU(4).
		Set("1", 1).
		Set("2", 2).
		Set("3", 3).
		Set("4", 4)
	lru.ll.Foreach(func(node *LinkedListNode) bool {
		t.Log(node)
		return true
	})
	t.Log(">>>")

	lru.Set("5", 5)
	lru.ll.Foreach(func(node *LinkedListNode) bool {
		t.Log(node)
		return true
	})
	t.Log(">>>")

	t.Log(lru.Get("2"))
	lru.ll.Foreach(func(node *LinkedListNode) bool {
		t.Log(node)
		return true
	})
	t.Log(">>>")

	lru.Set("6", 6)
	lru.ll.Foreach(func(node *LinkedListNode) bool {
		t.Log(node)
		return true
	})
	t.Log(">>>")
}
