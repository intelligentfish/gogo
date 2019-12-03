package xredis

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"sync"
	"time"
)

// Mutex Redis锁
type Mutex struct {
	key    string             // Mutex Key
	uuid   string             // Mutex UUID
	ms     int                // 超时间隔
	ch     chan<- bool        // 加锁结果回调
	locked bool               // 是否锁住
	wg     sync.WaitGroup     // 等待组
	ctx    context.Context    // 上下文
	cancel context.CancelFunc // 撤销方法
}

// lock 加锁
func (object *Mutex) lock() {
	value, err := GetRedisClientInstance().C.Eval(`
local value = redis.call('get', KEYS[1])
if not value then
	redis.call('set', KEYS[1], ARGV[1], 'px', ARGV[2])
	return 1
elseif value ~= ARGV[1] then
	return 0
end
`, []string{object.key}, []string{object.uuid, fmt.Sprint(object.ms)}).Int()
	if nil != err {
		glog.Error(err)
		return
	}
	object.locked = 1 == value
	object.ch <- object.locked
}

// unlock 解锁
func (object *Mutex) unlock() {
	_, err := GetRedisClientInstance().C.Eval(`
local value = redis.call('get', KEYS[1])
if not value then
	return 1
elseif value ~= ARGV[1] then
	return 1
else
	redis.call('del', KEYS[1])
end
`, []string{object.key}, []string{object.uuid}).Int()
	if nil != err {
		glog.Error(err)
		return
	}
}

// loop 循环
func (object *Mutex) loop() {
	object.lock()
	d := time.Duration(object.ms/5) * time.Millisecond
	if 0 == d {
		d = 1
	}
	t := time.NewTimer(d)
	var value int
	var err error
loop:
	for {
		select {
		case <-t.C:
			if value, err = GetRedisClientInstance().C.Eval(`
local value = redis.call('get', KEYS[1])
if not value then
	redis.call('set', KEYS[1], ARGV[1], 'px', ARGV[2])
	return 1
elseif value ~= ARGV[1] then
	return 0
else
	redis.call('set', KEYS[1], ARGV[1], 'px', ARGV[2], 'xx')
	return 2
end
`, []string{object.key}, []string{object.uuid, fmt.Sprint(object.ms)}).Int(); nil != err {
				glog.Error(err)
			} else {
				old := object.locked
				object.locked = 0 != value
				if old != object.locked {
					object.ch <- object.locked
				}
			}
			t.Reset(d)
		case <-object.ctx.Done():
			break loop
		}
	}
	t.Stop()
	object.unlock()
	object.wg.Done()
}

// Unlock解锁
func (object *Mutex) Unlock() {
	object.cancel()
	object.wg.Wait()
}

// NewRedisMutex 工厂方法
func NewRedisMutex(key, uuid string,
	ms int, /*ms*/
	ch chan<- bool) *Mutex {
	object := &Mutex{
		key:  key,
		uuid: uuid,
		ms:   ms,
		ch:   ch,
	}
	object.ctx, object.cancel = context.WithCancel(context.Background())
	object.wg.Add(1)
	go object.loop()
	return object
}
