package xetcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"sync/atomic"
	"time"
)

// etcd扩展
type XETCD struct {
	isClosed int32
	c        *clientv3.Client
}

// NewXETCD 工厂方法
func NewXETCD(cfg *clientv3.Config) (x *XETCD, err error) {
	x = &XETCD{}
	x.c, err = clientv3.New(*cfg)
	return
}

// withCancel 获取取消
func (object *XETCD) withCancel(timeout time.Duration, callback func(ctx context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	callback(ctx)
	cancel()
}

// IsClosed 是否已经关闭
func (object *XETCD) IsClosed() bool {
	return 1 == atomic.LoadInt32(&object.isClosed)
}

// Close 关闭
func (object *XETCD) Close() {
	if nil == object.c || !atomic.CompareAndSwapInt32(&object.isClosed, 0, 1) {
		return
	}
	object.c.Close()
}

// GetClient 获取Client
func (object *XETCD) GetClient() *clientv3.Client {
	return object.c
}

// PutWithTTL 放入
func (object *XETCD) PutWithTTL(key, value string,
	ttl int64,
	timeout time.Duration) (err error) {
	var resp *clientv3.LeaseGrantResponse
	object.withCancel(timeout, func(ctx context.Context) {
		resp, err = object.c.Grant(ctx, ttl)
	})
	if nil != err {
		return
	}
	object.withCancel(timeout, func(ctx context.Context) {
		_, err = object.c.Put(ctx, key, value, clientv3.WithLease(resp.ID))
	})
	return
}

// KeepAlive 保活
func (object *XETCD) KeepAlive(key, value string,
	ttl int64,
	timeout time.Duration) (err error) {
	var resp *clientv3.LeaseGrantResponse
	object.withCancel(timeout, func(ctx context.Context) {
		resp, err = object.c.Grant(ctx, ttl)
	})
	if nil != err {
		return
	}
	object.withCancel(timeout, func(ctx context.Context) {
		_, err = object.c.Put(ctx, key, value, clientv3.WithLease(resp.ID))
	})
	if nil != err {
		return
	}
	_, err = object.c.KeepAlive(context.TODO(), resp.ID)
	if nil != err {
		return
	}
	return
}

// Watch 监视
func (object *XETCD) Watch(ctx context.Context,
	key string,
	callback func(event *clientv3.Event)) {
	for resp := range object.c.Watch(ctx, key) {
		for _, e := range resp.Events {
			callback(e)
		}
	}
}

// WatchPrefix 监视前缀
func (object *XETCD) WatchPrefix(ctx context.Context,
	key string,
	callback func(event *clientv3.Event)) {
	for resp := range object.c.Watch(ctx, key, clientv3.WithPrefix()) {
		for _, e := range resp.Events {
			callback(e)
		}
	}
}
