package xetcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/priority_define"
	"github.com/intelligentfish/gogo/routine_pool"
	"reflect"
	"sync"
	"time"
)

// 选项
type Option interface{}

// ServiceRegisterDiscovery 服务注册发现
type ServiceRegisterDiscovery struct {
	xetcd         *XETCD
	servicePrefix string
}

// NewServiceRegisterDiscovery 工厂方法
func NewServiceRegisterDiscovery(xetcd *XETCD,
	servicePrefix string) *ServiceRegisterDiscovery {
	object := &ServiceRegisterDiscovery{
		xetcd:         xetcd,
		servicePrefix: servicePrefix,
	}
	return object
}

// shutdown 关闭
func (object *ServiceRegisterDiscovery) shutdown() {
	if !object.xetcd.IsClosed() {
		object.xetcd.Close()
	}
}

// RunAsServer 运行于服务状态
func (object *ServiceRegisterDiscovery) RunAsServer(onlineHook func(key, value []byte),
	offlineHook func(key []byte)) {
	var wg sync.WaitGroup
	wg.Add(1)
	routine_pool.GetInstance().PostTask(func(params []interface{}) interface{} {
		defer wg.Done()
		object.xetcd.WatchPrefix(context.TODO(), object.servicePrefix, func(event *clientv3.Event) {
			switch event.Type {
			case clientv3.EventTypePut:
				if nil != onlineHook {
					onlineHook(event.Kv.Key, event.Kv.Value)
				}
			case clientv3.EventTypeDelete:
				if nil != offlineHook {
					offlineHook(event.Kv.Key)
				}
			}
		})
		return nil
	}, "ServiceRegisterDiscovery")
	event_bus.GetInstance().Mounting(reflect.TypeOf(&event.AppShutdownEvent{}),
		func(param interface{}) {
			if priority_define.ServiceRegisterDiscoveryPriority !=
				param.(*event.AppShutdownEvent).ShutdownPriority {
				return
			}
			object.shutdown()
			wg.Wait()

			glog.Info("ServiceRegisterDiscovery shutdown")
		})
}

// Register 注册
func (object *ServiceRegisterDiscovery) Register(name, config string) (err error) {
	key := fmt.Sprintf("%s:%s", object.servicePrefix, name)
	err = object.xetcd.KeepAlive(key, config, 1, 5*time.Second)
	return
}
