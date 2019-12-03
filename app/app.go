package app

import (
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/priority_define"
	"github.com/intelligentfish/gogo/routine_pool"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
)

const (
	ExitOK = iota // 退出码成功
)

var (
	once     sync.Once // 执行一次
	instance *App      // 程序实例
)

// App 应用
type App struct {
	PID        int
	isShutdown bool
	signalCh   chan os.Signal
}

// newApp 工厂方法
func newApp() *App {
	return &App{
		PID:      os.Getpid(),
		signalCh: make(chan os.Signal, 1),
	}
}

// notifyShutdown 通知关闭
func (object *App) notifyShutdown() {
	for priority := priority_define.ShutdownPriorityMax; priority >= priority_define.ShutdownPriorityUnknown; priority-- {
		glog.Errorf("notify priority: %d component shutdown", priority)
		event := &event.AppShutdownEvent{ShutdownPriority: priority}
		event_bus.GetInstance().SyncNotify(reflect.TypeOf(event), event)
	}
}

// Shutdown 关闭
func (object *App) Shutdown() {
	if nil != object.signalCh &&
		0 < cap(object.signalCh) &&
		len(object.signalCh) < cap(object.signalCh) {
		object.signalCh <- syscall.SIGQUIT
	}
}

// IsShutdown 是否已关闭
func (object *App) IsShutdown() bool {
	return object.isShutdown
}

// WaitShutdown 等待关闭
func (object *App) WaitShutdown() {
	signal.Notify(object.signalCh)
	for s := range object.signalCh {
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTERM:
			glog.Errorf("signal: %v, shutdown", s)
			object.isShutdown = true
			object.notifyShutdown()
			close(object.signalCh)

			event_bus.GetInstance().Stop()    // 停止事件总线
			routine_pool.GetInstance().Stop() // 停止协程池
			os.Exit(ExitOK)                   // 退出进程

		default:
			glog.Errorf("signal: %v, not handled", s)
		}
	}
}

// GetInstance 获取单例
func GetInstance() *App {
	once.Do(func() {
		instance = newApp()
	})
	return instance
}
