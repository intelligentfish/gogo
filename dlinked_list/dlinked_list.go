package dlinked_list

import (
	"bytes"
	"fmt"
	"github.com/intelligentfish/gogo/comparable"
)

// DLinkedListNode 双向链表节点
type DLinkedListNode struct {
	Data comparable.Comparable // 数据
	Next *DLinkedListNode      // 后一个节点
	Prev *DLinkedListNode      // 前一个节点
}

// DLinkedList 双向链表
type DLinkedList struct {
	head   *DLinkedListNode
	tail   *DLinkedListNode
	length int
}

// NewDLinkedList 工厂方法
func NewDLinkedList() *DLinkedList {
	return &DLinkedList{
		head:   nil,
		tail:   nil,
		length: 0,
	}
}

// Add 添加节点
func (object *DLinkedList) Add(data comparable.Comparable) {
	if nil == object.head && nil == object.tail {
		object.head = &DLinkedListNode{Data: data}
		object.tail = object.head
		object.head.Next = object.tail
		object.tail.Prev = object.head
	} else {
		node := &DLinkedListNode{Data: data}
		node.Prev = object.tail
		object.tail.Next = node
		object.tail = node
	}
	object.length++
}

// Find 查找节点
func (object *DLinkedList) Find(data interface{}) (node *DLinkedListNode) {
	return nil
}

// Remove 删除节点
func (object *DLinkedList) Remove(node *DLinkedListNode) {

}

// InsertBefore 之前插入节点
func (object *DLinkedList) InsertBefore(node, before *DLinkedListNode) {

}

// InsertAfter 之后插入节点
func (object *DLinkedList) InsertAfter(node, after *DLinkedListNode) {

}

// Head 头
func (object *DLinkedList) Head() *DLinkedListNode {
	return object.head
}

// Tail 尾
func (object *DLinkedList) Tail() *DLinkedListNode {
	return object.tail
}

// String 字符串描述
func (object *DLinkedList) String() string {
	var i int
	var buf bytes.Buffer
	fmt.Fprint(&buf, "[")
	for node := object.head; nil != node; node = node.Next {
		fmt.Fprint(&buf, node.Data)
		if i < object.length-1 {
			fmt.Fprint(&buf, ",")
		}
		i++
	}
	fmt.Fprint(&buf, "]")
	return buf.String()
}
