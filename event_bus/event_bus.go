package event_bus

import (
	"context"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/priority_define"
	"github.com/intelligentfish/gogo/routine_pool"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

// 常量
const (
	EventBusWorkerSize = 32   // 工作者大小
	NotifyChanMaxSize  = 1024 // 通知通道最大长度
)

var (
	once     sync.Once // 执行一次
	instance *EventBus // 单实例
)

// Notifiable 可通知接口
type Notifiable interface {
	// 获取Key
	GetKey() string
	// 通知
	Notify(param interface{})
}

// DefaultNotifiable 可通知接口默认实现
type DefaultNotifiable struct {
	Key      string                  // key
	callback func(param interface{}) // 回调
}

// NewDefaultNotifiable 工厂方法
func NewDefaultNotifiable(key string,
	callback func(param interface{})) *DefaultNotifiable {
	return &DefaultNotifiable{callback: callback}
}

// GetKey 获取Key
func (object *DefaultNotifiable) GetKey() string {
	return object.Key
}

// completed 通知方法
func (object *DefaultNotifiable) Notify(param interface{}) {
	if nil != object.callback {
		object.callback(param)
	}
}

// EventBeforeNotifyFilter 过滤器
type EventBeforeNotifyFilter func(eventType, event interface{}) bool

// NotifyParam 通知参数
type NotifyParam struct {
	NotifiableArray []Notifiable // 通知数据
	Param           interface{}  // 参数
}

// EventBus 事件总线
type EventBus struct {
	sync.RWMutex
	exitFlag                 int32                        // 退出标志
	ctx                      context.Context              // 上下文
	cancel                   context.CancelFunc           // 取消方法
	wg                       *sync.WaitGroup              // 运行组
	notifyCh                 chan *NotifyParam            // 通知通道
	eventGroup               map[interface{}][]Notifiable // 事件组
	eventBeforeNotifyFilters []EventBeforeNotifyFilter    // 通知过滤器
}

// newEventBus 工厂方法
func newEventBus() *EventBus {
	object := &EventBus{
		exitFlag:   0,
		wg:         &sync.WaitGroup{},
		notifyCh:   make(chan *NotifyParam, NotifyChanMaxSize),
		eventGroup: make(map[interface{}][]Notifiable, 0),
	}
	object.ctx, object.cancel = context.WithCancel(context.Background())
	object.start()
	return object
}

// InstallBeforeNotifyFilter 安装发送消息过滤器
func (object *EventBus) InstallBeforeNotifyFilter(filter EventBeforeNotifyFilter) {
	object.Lock()
	object.eventBeforeNotifyFilters = append(object.eventBeforeNotifyFilters, filter)
	object.Unlock()
}

// wait 等待事件
func (object *EventBus) wait() {
	defer func() {
		if r := recover(); nil != r {
			//打印调用栈
			debug.PrintStack()
			//退还资源
			object.wg.Done()
			//事件循环非正常退出
			routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
				object.wait()
				return nil
			}, "EventBus")
		}
	}()
	object.wg.Add(1)
loop:
	for {
		select {
		case <-object.ctx.Done():
			break loop
		case notifyParam := <-object.notifyCh:
			for _, notifiable := range notifyParam.NotifiableArray {
				notifiable.Notify(notifyParam.Param)
			}
		}
	}
	for 0 != len(object.notifyCh) {
		notifyParam := <-object.notifyCh
		for _, notifiable := range notifyParam.NotifiableArray {
			notifiable.Notify(notifyParam.Param)
		}
	}
	object.wg.Done()
}

// notify 通知事件
func (object *EventBus) notify(notifiableArray []Notifiable, param interface{}) {
	object.notifyCh <- &NotifyParam{
		NotifiableArray: notifiableArray,
		Param:           param,
	}
}

// start 运行
func (object *EventBus) start() {
	for i := 0; i < EventBusWorkerSize; i++ {
		routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
			object.wait()
			return nil
		}, "EventBus")
	}
}

// stop 停止
func (object *EventBus) Stop() {
	atomic.StoreInt32(&object.exitFlag, 1)
	object.cancel()
	object.wg.Wait()
	for 0 != len(object.notifyCh) {
		notifyParam := <-object.notifyCh
		for _, notifiable := range notifyParam.NotifiableArray {
			notifiable.Notify(notifyParam.Param)
		}
	}
	close(object.notifyCh)
}

// Mounting 挂载一次
func (object *EventBus) MountingOnce(event interface{},
	uniqueKey string,
	callback func(param interface{})) *EventBus {
	object.Lock()
	var notifiableArray []Notifiable
	if v, ok := object.eventGroup[event]; !ok {
		notifiableArray = make([]Notifiable, 0)
	} else {
		notifiableArray = v
	}
	found := false
	if 0 < len(uniqueKey) {
		for _, notifiable := range notifiableArray {
			if notifiable.GetKey() == uniqueKey {
				found = true
				break
			}
		}
	}
	if !found {
		notifiableArray = append(notifiableArray, NewDefaultNotifiable(uniqueKey, callback))
	}
	object.eventGroup[event] = notifiableArray
	object.Unlock()
	return object
}

// Mounting 挂载
func (object *EventBus) Mounting(event interface{},
	callback func(param interface{})) *EventBus {
	return object.MountingOnce(event, "", callback)
}

// completed 通知
func (object *EventBus) Notify(event, param interface{}) *EventBus {
	if 0 != atomic.LoadInt32(&object.exitFlag) {
		glog.Errorf("lost notify message: (%v, %v)", event, param)
		return object
	}

	//前置过滤
	object.RLock()
	for _, filter := range object.eventBeforeNotifyFilters {
		if !filter(event, param) {
			object.RUnlock()
			return object
		}
	}
	object.RUnlock()

	var notifiableArray []Notifiable
	object.RLock()
	if v, ok := object.eventGroup[event]; ok {
		notifiableArray = make([]Notifiable, len(v))
		copy(notifiableArray, v)
	}
	object.RUnlock()
	if nil != notifiableArray && 0 != len(notifiableArray) {
		object.notify(notifiableArray, param)
	}
	return object
}

// SyncNotify 同步通知
func (object *EventBus) SyncNotify(event, param interface{}) *EventBus {
	if 0 != atomic.LoadInt32(&object.exitFlag) {
		glog.Errorf("lost notify message: (%v, %v)", event, param)
		return object
	}

	//前置过滤
	object.RLock()
	for _, filter := range object.eventBeforeNotifyFilters {
		if !filter(event, param) {
			object.RUnlock()
			return object
		}
	}
	object.RUnlock()

	var notifiableArray []Notifiable
	object.RLock()
	if v, ok := object.eventGroup[event]; ok {
		notifiableArray = make([]Notifiable, len(v))
		copy(notifiableArray, v)
	}
	object.RUnlock()
	if nil != notifiableArray && 0 != len(notifiableArray) {
		for _, notifiable := range notifiableArray {
			notifiable.Notify(param)
		}
	}

	return object
}

// 通知所有组件关闭
func (object *EventBus) NotifyAllComponentShutdown() {
	for p := priority_define.ShutdownPriorityMax; p > priority_define.ShutdownPriorityUnknown; p-- {
		object.SyncNotify(reflect.TypeOf(&event.AppShutdownEvent{}), &event.AppShutdownEvent{p})
	}
}

// GetInstance 获取单例
func GetInstance() *EventBus {
	once.Do(func() {
		instance = newEventBus()
	})
	return instance
}
