package main

import (
	"context"
	"flag"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/app"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/routine_pool"
)

// mainImpl main实现
func mainImpl() {
}

// 入口
func main() {
	// 解析命令行参数
	flag.Parse()

	// 设置日志标志
	flag.Set("log_dir", "logs")
	flag.Set("alsologtostderr", "true")
	flag.Set("v", "0")

	// 输出进程PID
	glog.Infof("app id: %d", app.GetInstance().PID)

	// 提交实现方法
	routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
		mainImpl()
		return nil
	}, "mainImpl")

	// 注册关闭回调&等待程序退出
	app.GetInstance().AddShutdownHook(
		event_bus.GetInstance().NotifyAllComponentShutdown, // 通知所有组件关闭
		event_bus.GetInstance().Stop,                       // 停止事件总线
		routine_pool.GetInstance().Stop,                    // 停止协程池
	).WaitShutdown()
}
