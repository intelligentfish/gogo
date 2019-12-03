package block_chan

import (
	"bytes"
	"fmt"
	"github.com/intelligentfish/gogo/auto_lock"
	"github.com/intelligentfish/gogo/linked_list"
	"github.com/intelligentfish/gogo/util"
	"time"
)

// 区块链节点
type BlockChanNode struct {
	Index        int64
	Timestamp    int64
	Data         []byte
	PreviousHash []byte
	Hash         []byte
}

// NewBlockChanNode 工厂方法
func NewBlockChanNode(data []byte) *BlockChanNode {
	object := &BlockChanNode{
		Timestamp: time.Now().UnixNano(),
		Data:      data,
	}
	return object
}

// CalcHash 计算块Hash
func (object *BlockChanNode) CalcHash() (err error) {
	var buf bytes.Buffer
	fmt.Fprint(&buf,
		object.Index,
		object.Timestamp,
		object.Data,
		object.PreviousHash)
	object.Hash, err = util.ReaderSHA256(&buf)
	return
}

// CompareTo 比较
func (object *BlockChanNode) CompareTo(other interface{}) int {
	return int(object.Index - other.(*BlockChanNode).Index)
}

// 区块链
type BlockChan struct {
	auto_lock.AutoLock
	*linked_list.LinkedList
}

// NewBlockChan 工厂方法
func NewBlockChan() *BlockChan {
	object := &BlockChan{LinkedList: linked_list.NewLinkedList()}
	node := NewBlockChanNode(nil)
	node.Timestamp = time.Date(2019,
		11,
		26,
		0,
		0,
		0,
		0,
		time.UTC).UnixNano()
	util.PanicOnError(node.CalcHash())
	object.Add(node)
	return object
}

// AddBlock 添加区块
func (object *BlockChan) AddBlock(node *BlockChanNode) *BlockChan {
	object.WithLock(false, func() {
		previous := object.LinkedList.Tail().Data.(*BlockChanNode)
		node.Index = previous.Index + 1
		node.PreviousHash = previous.Hash
		util.PanicOnError(node.CalcHash())
		object.LinkedList.Add(node)
	})
	return object
}

// Valid 校验
func (object *BlockChan) Valid() bool {
	var ret bool
	var previous *linked_list.LinkedListNode
	object.WithLock(true, func() {
		object.LinkedList.Foreach(func(index int, node *linked_list.LinkedListNode) bool {
			if 0 == index {
				previous = node
				return true
			}
			blockChanNode := node.Data.(*BlockChanNode)
			hash := make([]byte, len(blockChanNode.Hash))
			copy(hash, blockChanNode.Hash)
			util.PanicOnError(blockChanNode.CalcHash())
			if !bytes.Equal(hash, blockChanNode.Hash) ||
				!bytes.Equal(previous.Data.(*BlockChanNode).Hash, blockChanNode.PreviousHash) {
				return false
			}
			previous = node
			if index == object.LinkedList.Length()-1 {
				ret = true
			}
			return true
		})
	})
	return ret
}
