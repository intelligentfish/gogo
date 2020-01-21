package lru

import "fmt"

// LinkedListNode 双向链表节点
type LinkedListNode struct {
	Prev *LinkedListNode // 前驱节点
	Next *LinkedListNode // 后继节点
	Data interface{}     // 数据
}

// String 字符串描述
func (object *LinkedListNode) String() string {
	return fmt.Sprintf("node{Data:%v}", object.Data)
}

// LinkedList 双向链表
type LinkedList struct {
	Header *LinkedListNode // 链表头
	Tail   *LinkedListNode // 链表尾
}

// NewLinkedList 工厂方法
func NewLinkedList() *LinkedList {
	return &LinkedList{}
}

// AddFirst 添加到链首
func (object *LinkedList) AddFirst(node *LinkedListNode) *LinkedList {
	if nil == object.Header {
		object.Header = node
		object.Tail = node
	} else {
		node.Next = object.Header
		object.Header.Prev = node
		object.Header = node
	}
	return object
}

// AddLast 添加到链尾
func (object *LinkedList) AddLast(node *LinkedListNode) *LinkedList {
	if nil == object.Tail {
		object.Header = node
		object.Tail = node
	} else {
		object.Tail.Next = node
		node.Prev = object.Tail
		object.Tail = node
	}
	return object
}

// RemoveFirst 删除链首
func (object *LinkedList) RemoveFirst() (node *LinkedListNode) {
	if nil == object.Header {
		return
	}
	node = object.Header
	if node == object.Tail {
		object.Header = nil
		object.Tail = nil
		return
	}
	object.Header = node.Next
	object.Header.Prev = nil
	return
}

// RemoveLast 删除链尾
func (object *LinkedList) RemoveLast() (node *LinkedListNode) {
	if nil == object.Tail {
		return
	}
	node = object.Tail
	if node == object.Header {
		object.Header = nil
		object.Tail = nil
		return
	}
	object.Tail = node.Prev
	object.Tail.Next = nil
	return
}

// Remove 删除节点
func (object *LinkedList) Remove(node *LinkedListNode) {
	if nil == node {
		return
	}
	if object.Header == node {
		object.RemoveFirst()
	} else if object.Tail == node {
		object.RemoveLast()
	} else {
		node.Prev.Next = node.Next
		node.Next.Prev = node.Prev
	}
}

// Foreach 迭代
func (object *LinkedList) Foreach(callback func(node *LinkedListNode) bool) {
	cur := object.Header
	for nil != cur && callback(cur) {
		cur = cur.Next
	}
}

// LRU LRU缓存
type LRU struct {
	capacity int                             // 容量
	cache    map[interface{}]*LinkedListNode // 缓存
	ll       *LinkedList                     // 双向链表
}

// NewLRU 工厂方法
func NewLRU(capacity int) *LRU {
	return &LRU{
		capacity: capacity,
		cache:    make(map[interface{}]*LinkedListNode),
		ll:       NewLinkedList(),
	}
}

// Get 获取缓存
func (object *LRU) Get(key interface{}) (value interface{}, ok bool) {
	var node *LinkedListNode
	node, ok = object.cache[key]
	if ok {
		value = node.Data
		object.ll.Remove(node)
		object.ll.AddFirst(node)
	}
	return
}

// Set 设置缓存
func (object *LRU) Set(key, value interface{}) *LRU {
	if object.capacity <= len(object.cache) {
		node := object.ll.RemoveLast()
		if nil != node {
			delete(object.cache, node)
		}
	}
	node := &LinkedListNode{Data: value}
	object.cache[key] = node
	object.ll.AddLast(node)
	return object
}
