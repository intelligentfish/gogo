package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/app"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/routine_pool"
)

// business 业务
func business() {
	//TODO
}

// 入口
func main() {
	flag.Parse()
	glog.Infof("app id: %d", app.GetInstance().PID)
	business()
	app.GetInstance().AddShutdownHook(
		event_bus.GetInstance().NotifyAllComponentShutdown, // 通知所有组件关闭
		event_bus.GetInstance().Stop,                       // 停止事件总线
		routine_pool.GetInstance().Stop,                    // 停止协程池
	).WaitShutdown()
}
