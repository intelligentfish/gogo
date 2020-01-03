package byte_buf

import "sync"

var (
	poolOnce     sync.Once
	poolInstance *Pool
)

// Pool 池
type Pool struct {
	*sync.Pool
}

// newPool 工厂方法
func newPool() *Pool {
	return &Pool{
		Pool: &sync.Pool{New: func() interface{} {
			return &ByteBuf{}
		}},
	}
}

// Borrow 借缓存
func (object *Pool) Borrow(options ...ByteBufOption) *ByteBuf {
	return object.Get().(*ByteBuf).Initialize(options...)
}

// Return 还缓存
func (object *Pool) Return(buffer *ByteBuf) {
	object.Pool.Put(buffer)
}

// GetPoolInstance 获取池实例
func GetPoolInstance() *Pool {
	poolOnce.Do(func() {
		poolInstance = newPool()
	})
	return poolInstance
}
