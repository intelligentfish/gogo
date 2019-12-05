package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/app"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/priority_define"
	"github.com/intelligentfish/gogo/routine_pool"
	"reflect"
)

func main() {
	flag.Parse()
	glog.Infof("app id: %d", app.GetInstance().PID)
	// TODO
	app.GetInstance().AddShutdownHook(func() {
		// 通知所有组件关闭
		for priority := priority_define.ShutdownPriorityMax; priority >= priority_define.ShutdownPriorityUnknown; priority-- {
			glog.Errorf("notify priority: %d component shutdown", priority)
			event := &event.AppShutdownEvent{ShutdownPriority: priority}
			event_bus.GetInstance().SyncNotify(reflect.TypeOf(event), event)
		}
		event_bus.GetInstance().Stop()    // 停止事件总线
		routine_pool.GetInstance().Stop() // 停止协程池
	}).WaitShutdown()
}
