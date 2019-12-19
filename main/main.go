package main

import (
	"context"
	"flag"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/app"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/routine_pool"
	"io"
	"os"
	"os/exec"
)

// business 业务
func business(asChild bool) {
	if asChild {
		routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
			readBuf := make([]byte, 1<<12)
			pipe := os.NewFile(uintptr(3), "PIPE")
			defer pipe.Close()
		loop:
			for {
				_, err := pipe.Read(readBuf)
				if nil != err {
					glog.Error(err)
					break
				}
				var cmd string
				for i := 0; i < len(readBuf)-4; i++ {
					if '\r' == readBuf[i] && '\n' == readBuf[i+1] && '\r' == readBuf[i+2] && '\n' == readBuf[i+3] {
						cmd = string(readBuf[0:i])
						break
					}
				}
				switch cmd {
				case "EXIT":
					app.GetInstance().Shutdown()
					break loop
				case "UPDATE":
				}
			}
			return nil
		}, "BusinessReader")
	}
}

// mainImpl main实现
func mainImpl() {
	// asyncCopyStream 异步复制流
	asyncCopyStream := func(dst io.Writer, src io.Reader, name string) {
		routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
		loop:
			for {
				select {
				case <-ctx.Done():
					break loop
				default:
					if _, err := io.Copy(dst, src); nil != err {
						glog.Error(err)
						break loop
					}
				}
			}
			return nil
		}, name)
	}
	if _, ok := app.GetInstance().ArgsMap["child"]; ok {
		business(true)
		return
	}
	if _, ok := app.GetInstance().ArgsMap["daemon"]; ok {
		// 附加启动参数
		rFile, wFile, err := os.Pipe()
		if nil != err {
			glog.Error(err)
			return
		}
		os.Args = append(os.Args, "child")
		cmdObject := exec.Command(os.Args[0], os.Args[1:]...)
		cmdObject.ExtraFiles = append(cmdObject.ExtraFiles, rFile)
		// 标准输出流pipe
		childStdOut, err := cmdObject.StdoutPipe()
		if nil != err {
			glog.Error(err)
			return
		}
		defer childStdOut.Close()
		// 标准错误流pipe
		childStdErr, err := cmdObject.StderrPipe()
		if nil != err {
			glog.Error(err)
			return
		}
		defer childStdErr.Close()
		// 读取子进程标准输出流
		asyncCopyStream(os.Stdout, childStdOut, "ChildStdOutReader")
		// 读取子进程标准错误流
		asyncCopyStream(os.Stderr, childStdErr, "ChildStdErrReader")
		// 启动子进程
		err = cmdObject.Start()
		if nil != err {
			glog.Error(err)
			return
		}
		// 退出
		app.GetInstance().AddShutdownHook(func() {
			if _, err = wFile.Write([]byte("EXIT\r\n\r\n")); nil != err {
				glog.Error(err)
			}
		})
		// 等待子进程结束
		if err = cmdObject.Wait(); nil != err {
			glog.Error(err)
			return
		}
		return
	}
	business(false)
}

// 入口
func main() {
	// 解析命令行参数
	flag.Parse()
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
