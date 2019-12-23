package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/app"
	"github.com/intelligentfish/gogo/daemon"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/priority_define"
	"github.com/intelligentfish/gogo/routine_pool"
	"net"
	"net/http"
	"os"
	"reflect"
	"time"
)

// mainImpl main实现
func mainImpl(tcpFds map[string]int, ready chan bool) {
	// 输出进程PID
	glog.Infof("app id: %d", app.GetInstance().PID)

	// 提交实现方法
	webFD := tcpFds["web"]
	webLn, err := net.FileListener(os.NewFile(uintptr(webFD), "web"))
	if nil != err {
		glog.Error(err)
		return
	}

	// 模拟耗时操作...
	time.Sleep(10 * time.Second)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.GET("/api/v1/pid", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, fmt.Sprintf("pid: %d, ppid: %d\n", os.Getpid(), os.Getppid()))
	})
	routine_pool.GetInstance().CommitTask(func(ctx context.Context, params []interface{}) interface{} {
		if err = engine.RunListener(webLn); nil != err {
			glog.Error(err)
		}
		return nil
	}, "mainImpl")

	// 准备好
	ready <- true

	// 注册关闭事件
	event_bus.GetInstance().Mounting(reflect.TypeOf(&event.AppShutdownEvent{}),
		func(ctx context.Context, param interface{}) {
			if priority_define.HTTPServiceShutdownPriority !=
				param.(*event.AppShutdownEvent).ShutdownPriority {
				return
			}
			webLn.Close()
			glog.Error("web shutdown")
		})

	// 注册关闭回调&等待程序退出
	app.GetInstance().AddShutdownHook(
		event_bus.GetInstance().NotifyAllComponentShutdown, // 通知所有组件关闭
		event_bus.GetInstance().Stop,                       // 停止事件总线
		routine_pool.GetInstance().Stop,                    // 停止协程池
	).WaitShutdown() // 等待结束
}

// 入口
func main() {
	// 创建日志目录
	d := daemon.Default()
	d.Bootstrap(map[string]int{"web": 10080}, mainImpl, make(chan bool, 1))
}
