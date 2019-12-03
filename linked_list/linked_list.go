package linked_list

import (
	"bytes"
	"fmt"
	"github.com/intelligentfish/gogo/comparable"
)

// LinkedListNode 节点
type LinkedListNode struct {
	Data comparable.Comparable // 数据
	Next *LinkedListNode       // 下一个节点
}

// LinkedList 链表
type LinkedList struct {
	head   *LinkedListNode
	tail   *LinkedListNode
	length int
}

// NewLinkedList 工厂方法
func NewLinkedList() *LinkedList {
	return &LinkedList{}
}

// IsEmpty 是否为空
func (object *LinkedList) IsEmpty() bool {
	return object.head == object.tail && nil == object.head
}

// Length 长度
func (object *LinkedList) Length() int {
	return object.length
}

// Foreach 迭代
func (object *LinkedList) Foreach(callback func(index int, node *LinkedListNode) bool) {
	index := 0
	for start := object.head; nil != start; start = start.Next {
		if !callback(index, start) {
			break
		}
		index++
	}
}

// Add 添加
func (object *LinkedList) Add(data comparable.Comparable) *LinkedList {
	if object.IsEmpty() {
		object.tail = &LinkedListNode{Data: data}
		object.head = object.tail
	} else {
		object.tail.Next = &LinkedListNode{Data: data}
		object.tail = object.tail.Next
	}
	object.length++
	return object
}

// Remove 删除
func (object *LinkedList) Remove(data comparable.Comparable) (node *LinkedListNode) {
	var prev *LinkedListNode
	for start := object.head; nil != start; {
		if 0 == start.Data.CompareTo(data) {
			if object.head == start {
				object.head = start.Next
			} else if object.tail == start {
				prev.Next = nil
				object.tail = prev
			} else {
				prev.Next = start.Next
			}
			object.length--
			return start
		}
		prev = start
		start = start.Next
	}
	return nil
}

// Header 链首
func (object *LinkedList) Header() *LinkedListNode {
	return object.head
}

// Tail 链尾
func (object *LinkedList) Tail() *LinkedListNode {
	return object.tail
}

// 字符串描述
func (object *LinkedList) String() string {
	var buf bytes.Buffer
	fmt.Fprint(&buf, "LinkedList(")
	object.Foreach(func(index int, node *LinkedListNode) bool {
		fmt.Fprint(&buf, node.Data)
		if object.length-1 > index {
			fmt.Fprint(&buf, ",")
		}
		return true
	})
	fmt.Fprint(&buf, ")")
	return buf.String()
}
