package health_checker

import (
	"context"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/auto_lock"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/priority_define"
	"github.com/intelligentfish/gogo/routine_pool"
	"reflect"
	"sync"
	"time"
)

var (
	healthCheckerOnce     sync.Once      // 执行一次
	healthCheckerInstance *HealthChecker // 单实例
)

// healthCheckComponent 健康检查组件
type healthCheckComponent struct {
	UniqueName      string    // 名称
	OK              bool      // 成功
	LastRequestTime time.Time // 最后请求时间
	LastReplyTimes  time.Time // 最后回应时间
	MaxFailedTimes  int       // 最大失败次数
}

// HealthChecker 健康检查器
type HealthChecker struct {
	auto_lock.AutoLock
	checkInterval           time.Duration
	replyCostThreshold      time.Duration
	maxFailedTimesThreshold int
	healthCheckComponents   map[string]*healthCheckComponent
}

// NewHealthChecker 工厂方法
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		healthCheckComponents: make(map[string]*healthCheckComponent, 0),
	}
}

// SetCheckInterval 设置检查间隔
func (object *HealthChecker) SetCheckInterval(checkInterval time.Duration) *HealthChecker {
	object.WithLock(false, func() {
		object.checkInterval = checkInterval
	})
	return object
}

// getCheckInterval 获取检查间隔
func (object *HealthChecker) getCheckInterval() time.Duration {
	var d time.Duration
	object.WithLock(true, func() {
		d = object.checkInterval
	})
	if 0 >= d {
		d = 5 * time.Second
	}
	return d
}

// SetReplyCostThreshold 设置回执花费阈值
func (object *HealthChecker) SetReplyCostThreshold(threshold time.Duration) *HealthChecker {
	object.WithLock(false, func() {
		object.replyCostThreshold = threshold
	})
	return object
}

// getReplyCostThreshold 获取回执花费阈值
func (object *HealthChecker) getReplyCostThreshold() time.Duration {
	var t time.Duration
	object.WithLock(true, func() {
		t = object.replyCostThreshold
	})
	if 0 >= t {
		t = 10 * time.Second
	}
	return t
}

// SetMaxFailedTimesThreshold 设置最大失败次数阈值
func (object *HealthChecker) SetMaxFailedTimesThreshold(threshold int) *HealthChecker {
	object.WithLock(false, func() {
		object.maxFailedTimesThreshold = threshold
	})
	return object
}

// getMaxFailedTimesThreshold 获取最大失败次数阈值
func (object *HealthChecker) getMaxFailedTimesThreshold() int {
	var t int
	object.WithLock(true, func() {
		t = object.maxFailedTimesThreshold
	})
	if 0 >= t {
		t = 5
	}
	return t
}

// AddComponent 添加组件
func (object *HealthChecker) AddComponent(uniqueName string) *HealthChecker {
	object.WithLock(false, func() {
		object.healthCheckComponents[uniqueName] = &healthCheckComponent{
			UniqueName: uniqueName,
		}
	})
	return object
}

// Start 运行
func (object *HealthChecker) Start() {
	var wg sync.WaitGroup
	wg.Add(2)
	ctx, cancel := context.WithCancel(context.Background())
	// 延迟执行一次
	routine_pool.GetInstance().CommitTask(func(_ context.Context, params []interface{}) interface{} {
		defer wg.Done()
		select {
		case <-ctx.Done():
			break
		case <-time.After(object.getCheckInterval()):
			object.WithLock(false, func() {
				for _, v := range object.healthCheckComponents {
					v.LastRequestTime = time.Now()
				}
			})
			event_bus.GetInstance().Notify(reflect.TypeOf(&event.HealthCheckRequest{}),
				&event.HealthCheckRequest{})
		}
		return nil
	}, "HealthCheckerSender")
	// 检查
	routine_pool.GetInstance().CommitTask(func(_ context.Context, params []interface{}) interface{} {
		defer wg.Done()
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case <-time.After(1 * time.Second):
				object.WithLock(true, func() {
					for _, v := range object.healthCheckComponents {
						if v.LastRequestTime.IsZero() || v.LastReplyTimes.IsZero() {
							continue
						}
						if v.LastReplyTimes.After(v.LastRequestTime.Add(object.getReplyCostThreshold())) {
							//TODO 报告错误
							glog.Errorf("name: %s, health check timeout, request: %s, reply: %s",
								v.UniqueName,
								v.LastRequestTime,
								v.LastReplyTimes)
						}
						if v.MaxFailedTimes > object.getMaxFailedTimesThreshold() {
							//TODO 报告错误
							glog.Errorf("name: %s, health check failed times, request: %s, reply: %s, failed times: %d",
								v.UniqueName,
								v.LastRequestTime,
								v.LastReplyTimes,
								v.MaxFailedTimes)
						}
					}
				})
			}
		}
		return nil
	}, "HealthCheckerReporter")
	// 处理消息
	event_bus.GetInstance().Mounting(reflect.TypeOf(&event.AppShutdownEvent{}),
		func(ctx context.Context, param interface{}) {
			if priority_define.HealthCheckerPriority !=
				param.(*event.AppShutdownEvent).ShutdownPriority {
				return
			}
			cancel()
			wg.Wait()
			glog.Info("HealthChecker shutdown")
		}).Mounting(reflect.TypeOf(&event.HealthCheckReply{}),
		func(ctx context.Context, param interface{}) {
			reply := param.(*event.HealthCheckReply)
			object.WithLock(false, func() {
				v, ok := object.healthCheckComponents[reply.UniqueName]
				if !ok {
					v = &healthCheckComponent{
						UniqueName: reply.UniqueName,
					}
					object.healthCheckComponents[reply.UniqueName] = v
				}
				v.LastReplyTimes = time.Now()
				v.OK = reply.OK
				if !v.OK {
					v.MaxFailedTimes++
				}
			})
			routine_pool.GetInstance().CommitTask(func(ctx context.Context, params []interface{}) interface{} {
				select {
				case <-ctx.Done():
					break
				case <-time.After(object.getCheckInterval()):
					object.WithLock(false, func() {
						if v, ok := object.healthCheckComponents[reply.UniqueName]; ok {
							v.LastRequestTime = time.Now()
						}
					})
					event_bus.GetInstance().Notify(reflect.TypeOf(&event.HealthCheckRequest{}),
						&event.HealthCheckRequest{UniqueName: reply.UniqueName})
				}
				return nil
			}, "HealthCheckerSender")
		})
}

// GetHealthCheckerInstance 获取单例
func GetHealthCheckerInstance() *HealthChecker {
	healthCheckerOnce.Do(func() {
		healthCheckerInstance = NewHealthChecker()
	})
	return healthCheckerInstance
}
