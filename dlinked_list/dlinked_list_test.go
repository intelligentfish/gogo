package dlinked_list

import (
	"fmt"
	"testing"
)

type NodeData struct {
	Data int
}

func (object *NodeData) String() string {
	return fmt.Sprint(object.Data)
}

func (object *NodeData) CompareTo(other interface{}) int {
	if object == other {
		return 0
	}
	return object.Data - other.(*NodeData).Data
}

func TestDLinkedList(t *testing.T) {
	dLinkedList := NewDLinkedList()
	for i := 0; i < 10; i++ {
		dLinkedList.Add(&NodeData{Data: i})
	}
	t.Log(dLinkedList)
}
