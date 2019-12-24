package spin_lock

import (
	"runtime"
	"sync/atomic"
)

// SpinLock 自旋锁
type SpinLock struct {
	flag int32 // 锁标志
}

// New 工厂方法
func New() *SpinLock {
	return &SpinLock{}
}

// Lock 加锁
func (object *SpinLock) Lock() {
	for !atomic.CompareAndSwapInt32(&object.flag, 0, 1) {
		runtime.Gosched()
	}
}

// Unlock 释放
func (object *SpinLock) Unlock() {
	atomic.StoreInt32(&object.flag, 0)
}
