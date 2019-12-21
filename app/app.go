package app

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/auto_lock"
	"github.com/intelligentfish/gogo/util"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	ExitOK     = iota // 退出码成功
	DefaultSep = "="  // 默认命令行分隔符
)

var (
	once     sync.Once // 执行一次
	instance *App      // 程序实例
)

// App 应用
type App struct {
	auto_lock.AutoLock
	PID           int
	ArgsMap       map[string][]string
	isShutdown    bool
	signalCh      chan os.Signal
	shutdownHooks []ShutdownHook
}

// 关闭钩子
type ShutdownHook func()

// newApp 工厂方法
func newApp() *App {
	object := &App{
		PID:      os.Getpid(),
		signalCh: make(chan os.Signal, 1),
	}
	object.ArgsMap = util.ArrayToMap(os.Args, DefaultSep)
	return object
}

// notifyShutdown 通知关闭
func (object *App) notifyShutdown() {
	object.WithLock(true, func() {
		for _, hook := range object.shutdownHooks {
			if nil != hook {
				hook()
			}
		}
	})
}

// AddShutdownHook 添加关闭钩子
func (object *App) AddShutdownHook(hook ...ShutdownHook) *App {
	object.WithLock(false, func() {
		object.shutdownHooks = append(object.shutdownHooks, hook...)
	})
	return object
}

// Shutdown 关闭
func (object *App) Shutdown() *App {
	if nil != object.signalCh &&
		0 < cap(object.signalCh) &&
		len(object.signalCh) < cap(object.signalCh) {
		object.signalCh <- syscall.SIGQUIT
	}
	return object
}

// WaitComplete 等待完成
func (object *App) WaitComplete() {
	<-object.signalCh
}

// IsShutdown 是否已关闭
func (object *App) IsShutdown() bool {
	return object.isShutdown
}

// WaitShutdown 等待关闭
func (object *App) WaitShutdown() {
	signal.Notify(object.signalCh)
loop:
	for s := range object.signalCh {
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTERM:
			glog.Errorf("signal: %v, shutdown", s)
			object.isShutdown = true
			object.notifyShutdown()
			break loop
		default:
			glog.Errorf("signal: %v, not handled", s)
		}
	}
	fmt.Println("App shutdown")
	glog.Info("App shutdown")
	close(object.signalCh)
	// 太粗暴，注释掉
	//os.Exit(ExitOK) // 退出进程
}

// GetInstance 获取单例
func GetInstance() *App {
	once.Do(func() {
		instance = newApp()
	})
	return instance
}
