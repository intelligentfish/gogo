package main

import (
	"context"
	"flag"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/app"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/routine_pool"
	"github.com/intelligentfish/gogo/util"
	"github.com/intelligentfish/gogo/xcmd"
	"os"
	"time"
)

// mainImpl main实现
func mainImpl() {
	argsMap := util.ArrayToMap(os.Args, "=")
	_, isChild := argsMap["child"]
	if !isChild {
		// 父进程
		glog.Info("in parent: ", os.Getpid())

		// 创建日志目录
		os.RemoveAll("logs")
		os.Mkdir("logs", os.ModePerm)
		args := make([]string, len(os.Args))
		copy(args, os.Args)
		args = append(args, "child")

		xcmdObj := xcmd.New(args[0], args[1:]...)
		defer xcmdObj.Close()

		xcmdObj.Stdin = os.Stdin
		xcmdObj.Stdout = os.Stdout
		xcmdObj.Stderr = os.Stderr
		err := xcmdObj.Start()
		if nil != err {
			glog.Error(err)
			return
		}
		if err = xcmdObj.ParentRead(func(raw []byte) bool {
			request := string(raw)
			glog.Info("child request: ", request)
			return "OK" != request
		}); nil != err {
			glog.Error(err)
		}
		// 父进程完成退出时通知子进程退出
		routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
			<-ctx.Done()
			if err = xcmdObj.ParentWrite([]byte("EXIT")); nil != err {
				glog.Error(err)
			}
			return nil
		}, "ShutdownChild")
		if err = xcmdObj.Wait(); nil != err {
			glog.Error(err)
		}
		glog.Error("child exited")

		return
	}
	// 子进程

	glog.Info("in child: ", os.Getpid())

	xcmdObj := xcmd.FromFd(3, 4)
	defer xcmdObj.Close()

	// 模拟耗时操作
	time.Sleep(5 * time.Second)
	err := xcmdObj.ChildWrite([]byte("OK"))
	if nil != err {
		glog.Error(err)
		return
	}

	// 子进程读
	if err = xcmdObj.ChildRead(func(raw []byte) bool {
		request := string(raw)
		glog.Info("parent request: ", request)
		return "EXIT" != request
	}); nil != err {
		glog.Error(err)
	}
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
