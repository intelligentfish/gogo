package token_bucket

import (
	"context"
	"sync"
	"time"
)

// TokenBucket 令牌桶
type TokenBucket struct {
	bucket    chan interface{}
	rateLimit time.Duration
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewTokenBucket 工厂方法
func NewTokenBucket(ops int) *TokenBucket {
	object := &TokenBucket{
		bucket:    make(chan interface{}, ops),
		rateLimit: time.Duration(1000000000 / ops),
	}
	object.ctx, object.cancel = context.WithCancel(context.Background())
	return object
}

// Start 启动
func (object *TokenBucket) Start() {
	object.wg.Add(1)
	go func() {
		for {
			select {
			case <-object.ctx.Done():
				break
			case <-time.After(object.rateLimit * time.Nanosecond):
				object.bucket <- nil
			}
		}
		object.wg.Done()
	}()
}

// Stop 停止
func (object *TokenBucket) Stop() {
	object.cancel()
	object.wg.Wait()
}

// WithToken 获取token
func (object *TokenBucket) WithToken(callback func()) {
	<-object.bucket
	callback()
}
