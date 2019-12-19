package main

import (
	"context"
	"flag"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/app"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/pipe"
	"github.com/intelligentfish/gogo/routine_pool"
	"github.com/intelligentfish/gogo/util"
	"os"
	"os/exec"
	"time"
)

// mainImpl main实现
func mainImpl() {
	argsMap := util.ArrayToMap(os.Args, "=")
	_, isChild := argsMap["child"]
	if !isChild {
		// 父进程

		// 创建日志目录
		os.RemoveAll("logs")
		os.Mkdir("logs", os.ModePerm)
		args := make([]string, len(os.Args))
		copy(args, os.Args)
		args = append(args, "child")
		cmdObj := exec.Command(args[0], args[1:]...)
		cmdObj.Stdout = os.Stdout
		cmdObj.Stderr = os.Stderr
		cmdObj.Stdin = os.Stdin
		p := pipe.NewPIPE()
		cmdObj.ExtraFiles = append(cmdObj.ExtraFiles, p.GetReadPipe())
		err := cmdObj.Start()
		if nil != err {
			glog.Error(err)
			return
		}
		go func() {
			time.Sleep(5 * time.Second)
			p.Write([]byte("EXIT"))
		}()
		cmdObj.Wait()
		glog.Error("child exited")
		return
	}
	// 子进程
	p := &pipe.PIPE{}
	p.SetReadPipe(os.NewFile(uintptr(3), "pipe"))
	err := p.Read(func(data []byte) bool {
		glog.Info("request: ", string(data))
		return "EXIT" != string(data)
	})
	if nil != err {
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
