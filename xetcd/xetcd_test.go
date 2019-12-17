package xetcd

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/intelligentfish/gogo/app"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/priority_define"
	"reflect"
	"testing"
	"time"
)

func TestServiceRegisterDiscovery(t *testing.T) {
	xetcd, err := NewXETCD(&clientv3.Config{
		Endpoints:   []string{"172.25.73.216:2379"},
		DialTimeout: 5 * time.Second,
	})
	if nil != err {
		t.Error(err)
		return
	}
	NewServiceRegisterDiscovery(xetcd, "service").
		RunAsServer(func(key, value []byte) {
			t.Log("online:", string(key), string(value))
		}, func(key []byte) {
			t.Log("offline:", string(key))
		})
	app.GetInstance().AddShutdownHook(func() {
		for p := priority_define.ShutdownPriorityMax; p > priority_define.ShutdownPriorityUnknown; p-- {
			event_bus.GetInstance().SyncNotify(reflect.TypeOf(&event.AppShutdownEvent{}),
				&event.AppShutdownEvent{
					ShutdownPriority: p,
				})
		}
	}).WaitShutdown()
}

func TestServiceRegisterDiscovery_Register(t *testing.T) {
	xetcd, err := NewXETCD(&clientv3.Config{
		Endpoints:   []string{"172.25.73.216:2379"},
		DialTimeout: 5 * time.Second,
	})
	if nil != err {
		t.Error(err)
		return
	}
	NewServiceRegisterDiscovery(xetcd, "service").
		Register("web1", `{"port":8080}`)
}
