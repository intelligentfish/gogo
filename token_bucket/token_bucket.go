package token_bucket

import (
	"context"
	"sync"
	"time"
)

// TokenBucket 令牌桶
type TokenBucket struct {
	bucket    chan interface{}
	max       int
	ops       int
	rateLimit time.Duration
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

type TokenBucketOption func(object *TokenBucket)

func TokenBucketOPSOption(ops int) TokenBucketOption {
	return func(object *TokenBucket) {
		object.ops = ops
	}
}

func TokenBucketMaxOption(max int) TokenBucketOption {
	return func(object *TokenBucket) {
		object.max = max
	}
}

// NewTokenBucket 工厂方法
func NewTokenBucket(options ...TokenBucketOption) *TokenBucket {
	object := &TokenBucket{}
	for _, option := range options {
		option(object)
	}
	if 0 >= object.ops {
		object.ops = 1
	}
	if 0 >= object.max {
		object.max = 1_000_000
	}
	object.bucket = make(chan interface{}, object.max)
	object.UpdateOPS(object.ops)
	object.ctx, object.cancel = context.WithCancel(context.Background())
	return object
}

// GetOPS 获取OPS
func (object *TokenBucket) GetOPS() int {
	return object.ops
}

// UpdateOPS 更新OPS
func (object *TokenBucket) UpdateOPS(ops int) {
	object.ops = ops
	object.rateLimit = time.Duration(1000000000 / ops)
}

// Start 启动
func (object *TokenBucket) Start() {
	object.wg.Add(1)
	go func() {
	loop:
		for {
			select {
			case <-object.ctx.Done():
				break loop
			case <-time.After(object.rateLimit * time.Nanosecond):
				if object.ops > len(object.bucket) && len(object.bucket) < cap(object.bucket) {
					object.bucket <- nil
				}
			}
		}
		object.wg.Done()
	}()
}

// Stop 停止
func (object *TokenBucket) Stop() {
	object.cancel()
	object.wg.Wait()
	close(object.bucket)
}

// WithToken 获取token
func (object *TokenBucket) WithToken(callback func()) {
	<-object.bucket
	callback()
}
