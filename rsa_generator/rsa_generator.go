package rsa_generator

import (
	"context"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/priority_define"
	"github.com/intelligentfish/gogo/routine_pool"
	"github.com/intelligentfish/gogo/xrsa"
	"reflect"
	"runtime"
	"sync"
)

// 全局变量
var (
	rsaGeneratorOnce     sync.Once     // 执行一次
	rsaGeneratorInstance *RSAGenerator // 生成器单实例
)

// RSAGenerator RSA生成器
type RSAGenerator struct {
	ch     chan *xrsa.RSA
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// newRSAGenerator 工厂方法
func newRSAGenerator(chanSize, rsaKeySize int) *RSAGenerator {
	object := &RSAGenerator{ch: make(chan *xrsa.RSA, chanSize)}
	object.ctx, object.cancel = context.WithCancel(context.Background())
	event_bus.GetInstance().Mounting(reflect.TypeOf(&event.AppShutdownEvent{}),
		func(param interface{}) {
			if priority_define.RSAGeneratorShutdownPriority !=
				param.(*event.AppShutdownEvent).ShutdownPriority {
				return
			}

			object.cancel()
			object.wg.Wait()
			close(object.ch)

			glog.Info("RSAGenerator done")
		})
	concurrency := runtime.NumCPU() / 2
	if 0 == concurrency {
		concurrency = 1
	}
	object.wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		routine_pool.GetInstance().PostTask(func(params []interface{}) interface{} {
			defer object.wg.Done()
		loop:
			for {
				select {
				case <-object.ctx.Done():
					break loop
				default:
					rsaObject := &xrsa.RSA{}
					err := rsaObject.MakeKey(rsaKeySize)
					if nil != err {
						glog.Error(err)
						continue
					}
					object.ch <- rsaObject
				}
			}
			return nil
		}, "RSAGenerator")
	}
	return object
}

// GetRSAObject 获取RSA对象
func (object *RSAGenerator) GetRSAObject() *xrsa.RSA {
	return <-object.ch
}

// GetRSAGeneratorInstance 单实例
func GetRSAGeneratorInstance() *RSAGenerator {
	rsaGeneratorOnce.Do(func() {
		rsaGeneratorInstance = newRSAGenerator(1024, 2048)
	})
	return rsaGeneratorInstance
}
