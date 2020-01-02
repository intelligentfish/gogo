package xetcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"sync/atomic"
	"time"
)

// XETCD ETCD客户端
type XETCD struct {
	isClosed int32            // 是否已经关闭
	C        *clientv3.Client // 客户端
}

// NewXETCD 工厂方法
func NewXETCD(cfg *clientv3.Config) (c *XETCD, err error) {
	c = &XETCD{}
	c.C, err = clientv3.New(*cfg)
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
	if nil == object.C || !atomic.CompareAndSwapInt32(&object.isClosed, 0, 1) {
		return
	}
	object.C.Close()
}

// GetClient 获取Client
func (object *XETCD) GetClient() *clientv3.Client {
	return object.C
}

// GetWithPrefix 获取满足前缀的所有
func (object *XETCD) GetWithPrefix(key string, timeout time.Duration) (values map[string][]byte, err error) {
	var resp *clientv3.GetResponse
	object.withCancel(timeout, func(ctx context.Context) {
		resp, err = object.C.Get(ctx, key, clientv3.WithPrefix())
	})
	if nil != err {
		return
	}
	if 0 < resp.Count {
		values = make(map[string][]byte, resp.Count)
		for _, kv := range resp.Kvs {
			values[string(kv.Key)] = kv.Value
		}
	}
	return
}

// PutWithTTL 放入
func (object *XETCD) PutWithTTL(key, value string,
	ttl int64,
	timeout time.Duration) (err error) {
	var resp *clientv3.LeaseGrantResponse
	object.withCancel(timeout, func(ctx context.Context) {
		resp, err = object.C.Grant(ctx, ttl)
	})
	if nil != err {
		return
	}
	object.withCancel(timeout, func(ctx context.Context) {
		_, err = object.C.Put(ctx, key, value, clientv3.WithLease(resp.ID))
	})
	return
}

// KeepAlive 保活
func (object *XETCD) KeepAlive(key, value string,
	ttl int64,
	timeout time.Duration) (err error) {
	var resp *clientv3.LeaseGrantResponse
	object.withCancel(timeout, func(ctx context.Context) {
		resp, err = object.C.Grant(ctx, ttl)
	})
	if nil != err {
		return
	}
	object.withCancel(timeout, func(ctx context.Context) {
		_, err = object.C.Put(ctx, key, value, clientv3.WithLease(resp.ID))
	})
	if nil != err {
		return
	}
	_, err = object.C.KeepAlive(context.TODO(), resp.ID)
	if nil != err {
		return
	}
	return
}

// Watch 监视
func (object *XETCD) Watch(ctx context.Context,
	key string,
	callback func(event *clientv3.Event)) {
	for resp := range object.C.Watch(ctx, key) {
		for _, e := range resp.Events {
			callback(e)
		}
	}
}

// WatchPrefix 监视前缀
func (object *XETCD) WatchPrefix(ctx context.Context,
	key string,
	callback func(event *clientv3.Event)) {
	for resp := range object.C.Watch(ctx, key, clientv3.WithPrefix()) {
		for _, e := range resp.Events {
			callback(e)
		}
	}
}
