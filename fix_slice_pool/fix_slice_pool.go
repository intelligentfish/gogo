package fix_slice_pool

import (
	"github.com/intelligentfish/gogo/auto_lock"
	"sync"
)

var (
	fixSlicePoolOnce     sync.Once     // 执行一次
	fixSlicePoolInstance *FixSlicePool // 单实例
)

// 固定大小切片池
type FixSlicePool struct {
	auto_lock.AutoLock
	pool map[int]*sync.Pool
}

// 工厂方法
func newFixSlicePool() *FixSlicePool {
	return &FixSlicePool{pool: make(map[int]*sync.Pool)}
}

// 借切片
func (object *FixSlicePool) BorrowSlice(size int) (slice []byte) {
	object.WithLock(true, func() {
		v, ok := object.pool[size]
		if ok {
			slice = v.Get().([]byte)
		}
	})
	if 0 < len(slice) {
		return
	}
	object.WithLock(false, func() {
		object.pool[size] = &sync.Pool{
			New: func() interface{} {
				return make([]byte, size)
			},
		}
		slice = object.pool[size].Get().([]byte)
	})
	return
}

// 还切片
func (object *FixSlicePool) ReturnSlice(size int, slice []byte) {
	object.WithLock(false, func() {
		if v, ok := object.pool[size]; ok {
			v.Put(slice)
		}
	})
}

// GetFixSlicePoolInstance 获取单实例
func GetFixSlicePoolInstance() *FixSlicePool {
	fixSlicePoolOnce.Do(func() {
		fixSlicePoolInstance = newFixSlicePool()
	})
	return fixSlicePoolInstance
}
