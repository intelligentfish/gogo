package xmq

import (
	"context"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/priority_define"
	"github.com/intelligentfish/gogo/routine_pool"
	"github.com/streadway/amqp"
	"reflect"
	"sync"
	"time"
)

var (
	once     sync.Once       // 执行一次
	instance *RabbitMQClient // 单实例
)

// RabbitMQClient 支持重连的RabbitMQ客户端封装
type RabbitMQClient struct {
	url             string           // URL
	enableReconnect bool             // 启用重新连接
	Conn            *amqp.Connection // 连接
}

// newRabbitMQClient 工厂方法
func newRabbitMQClient() *RabbitMQClient {
	return &RabbitMQClient{
		enableReconnect: true,
	}
}

// waitDisconnectAndReconnect 等待断线并重连
func (object *RabbitMQClient) waitDisconnectAndReconnect() {
	receiver := make(chan *amqp.Error, 1)
	object.Conn.NotifyClose(receiver)
	routine_pool.GetInstance().CommitTask(func(ctx context.Context, params []interface{}) interface{} {
		var closed bool
		select {
		case _, closed = <-receiver:
		}
		if !closed {
			close(receiver)
		}
		time.Sleep(100 * time.Millisecond) // 避免CPU过高
		object.initialize(true)
		return nil
	}, "RabbitMQClient")
}

// initialize 初始化
func (object *RabbitMQClient) initialize(force bool) (err error) {
	object.Conn, err = amqp.Dial(object.url)
	if nil == err {
		glog.Infof("RabbitMQClient initialize ok")
	}
	if force {
		// 连接没有建立
		if nil == object.Conn {
			routine_pool.GetInstance().CommitTask(func(ctx context.Context, params []interface{}) interface{} {
				time.Sleep(100 * time.Millisecond)
				object.initialize(true)
				return nil
			}, "RabbitMQClient")
			return
		}

		event_bus.GetInstance().MountingOnce(reflect.TypeOf(&event.AppShutdownEvent{}),
			"RabbitMQClient",
			func(ctx context.Context, param interface{}) {
				if priority_define.MQShutdownPriority !=
					param.(*event.AppShutdownEvent).ShutdownPriority {
					return
				}
				if nil != object.Conn {
					object.Conn.Close()
					object.Conn = nil
				}
				glog.Info("RabbitMQClient done")
			})
		// 重连
		object.waitDisconnectAndReconnect()
	}
	return
}

// InitializeWithTimeout 初始化
func (object *RabbitMQClient) Initialize(url string) (err error) {
	object.url = url
	err = object.initialize(false)
	if nil == err {
		event_bus.GetInstance().MountingOnce(reflect.TypeOf(&event.AppShutdownEvent{}),
			"RabbitMQClient",
			func(ctx context.Context, param interface{}) {
				if priority_define.MQShutdownPriority !=
					param.(*event.AppShutdownEvent).ShutdownPriority {
					return
				}
				if nil != object.Conn {
					object.Conn.Close()
					object.Conn = nil
				}
				glog.Info("RabbitMQClient done")
			})
		// 重连
		object.waitDisconnectAndReconnect()
	}
	return
}

// 获取GetInstance单实例
func GetInstance() *RabbitMQClient {
	once.Do(func() {
		instance = newRabbitMQClient()
	})
	return instance
}
