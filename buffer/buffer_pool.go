package buffer

import "sync"

// 池
var Pool = sync.Pool{
	New: func() interface{} {
		return &Buffer{}
	},
}
