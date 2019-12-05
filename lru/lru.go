package lru

import "github.com/intelligentfish/gogo/linked_list"

// LRU LRU缓存
type LRU struct {
	maxSize int                         // 最大大小
	m       map[interface{}]interface{} // 缓存
	l       linked_list.LinkedList      // 链表
}

// NewLRU 工厂方法
func NewLRU(maxSize int) *LRU {
	return &LRU{
		maxSize: maxSize,
		m:       make(map[interface{}]interface{}),
	}
}

// tidy 整理
func (object *LRU) tidy(key interface{}) {
	for i := object.l.Header(); nil != i.Next; i = i.Next {

	}
}

// Get 获取
func (object *LRU) Get(key interface{}) (value interface{}) {
	v, ok := object.m[key]
	if !ok {
		return nil
	}
	object.tidy(key)
	return v
}

// Set 设置
func (object *LRU) Set(key, value interface{}) {
	object.m[key] = value
	object.tidy(key)
}
