package auto_lock

import (
	"github.com/golang/glog"
	"sync"
)

// AutoLock 自动锁
type AutoLock struct {
	sync.RWMutex
}

// WithLock 使用
func (object *AutoLock) WithLock(read bool, callback func()) {
	if read {
		object.RLock()
	} else {
		object.Lock()
	}
	callback()
	if read {
		object.RUnlock()
	} else {
		object.Unlock()
	}
}

// WithLockDebug
func (object *AutoLock) WithDebugLock(read bool, funcName string, callback func()) {
	if read {
		glog.Info(funcName, "RLock")
		object.RLock()
	} else {
		glog.Info(funcName, "Lock")
		object.Lock()
	}
	callback()
	if read {
		glog.Info(funcName, "RUnLock")
		object.RUnlock()
	} else {
		glog.Info(funcName, "UnLock")
		object.Unlock()
	}
}
