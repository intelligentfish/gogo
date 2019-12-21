package daemon

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/intelligentfish/gogo/app"
	"github.com/intelligentfish/gogo/routine_pool"
	"github.com/intelligentfish/gogo/xcmd"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/golang/glog"
)

// 响应
const (
	ReadyOK     = "ReadyOK"
	ReadyError  = "ReadyError"
	ExitRequest = "Exit"
	ExitReply   = "Exit"
)

// panicOnError 错误崩溃
func panicOnError(err error) {
	if nil != err {
		panic(err)
	}
}

// Daemon 守护进程
type Daemon struct {
	sync.RWMutex
	rebootTimes     int            // 最大重启次数
	upgradeFlag     int32          // 正常更新标志
	killedFlag      int32          // 正常停服标志
	origArgs        []string       // 程序原始运行参数
	wg              sync.WaitGroup // 等待组
	xCmdObj         *xcmd.XCmd     // 扩展Cmd
	childCmd        string         // 运行子进程命令 --child
	upgradeCmd      string         // 更新命名 --upgrade
	bootstrapArgs   string         // 引导参数 --bootstrap_args
	bootstrapLogDir string         // 引导日志
	pidFile         string         // PID文件
	tcpPorts        map[string]int // 业务逻辑层需要用的端口
}

// New 工厂方法
func New(childCmd, upgradeCmd, bootstrapArgs, bootstrapLogDir, pidFile string) *Daemon {
	return &Daemon{
		rebootTimes:     3,
		childCmd:        childCmd,
		upgradeCmd:      upgradeCmd,
		bootstrapArgs:   bootstrapArgs,
		bootstrapLogDir: bootstrapLogDir,
		pidFile:         pidFile,
	}
}

// Default 默认实现
func Default() *Daemon {
	return New("child",
		"upgrade",
		"bootstrap_args",
		"bootstrapLogs",
		"daemonPID")
}

// spawnChildProcess 生成孩子进程
func (object *Daemon) spawnChildProcess(tcpLnFiles map[string]*os.File) (xCmdObj *xcmd.XCmd, err error) {
	// 构建启动参数
	args := make([]string, len(object.origArgs))
	copy(args, object.origArgs)
	args = append(args, "--"+object.childCmd)

	// 构建XCmd
	xCmdObj = xcmd.New(args[0], args[1:]...)

	// 赋值标准流
	xCmdObj.Stdin = os.Stdin
	xCmdObj.Stdout = os.Stdout
	xCmdObj.Stderr = os.Stderr

	// 填入fd
	tcpLnFds := make(map[string]int)
	for k, f := range tcpLnFiles {
		tcpLnFds[k] = xCmdObj.AddFile(f).NextFd()
	}

	// 写入启动参数
	var raw []byte
	raw, err = json.Marshal(tcpLnFds)
	panicOnError(err)
	xCmdObj.Args = append(xCmdObj.Args,
		fmt.Sprintf("--%s=%s", object.bootstrapArgs, string(raw)))

	// 启动子进程
	if err = xCmdObj.Start(); nil != err {
		glog.Error(err)
		return
	}

	return
}

// replaceChildProcess 重启子进程
func (object *Daemon) replaceChildProcess(tcpLnFiles map[string]*os.File) (ok bool, err error) {
	object.Lock()
	defer object.Unlock()

	var newXCmdObj *xcmd.XCmd
	newXCmdObj, err = object.spawnChildProcess(tcpLnFiles)
	if nil != err {
		glog.Error(err)
		return
	}

	// 等待子进程启动成功
	ok = false
	if err = newXCmdObj.ParentRead(func(raw []byte) bool {
		if nil == raw || 0 >= len(raw) {
			return false
		}
		request := string(raw)
		glog.Info("child request: ", request)
		switch request {
		case ReadyOK:
			ok = true
			return false
		case ReadyError:
			return false
		default:
			return true
		}
	}); nil != err {
		glog.Error(err)
	}

	// 启动子进程失败
	if !ok {
		newXCmdObj.Close()
		newXCmdObj = nil
		return
	}

	if nil != object.xCmdObj {
		glog.Info("notify old child exit")
		// 通知退出
		if err = object.notifyChildExit(); nil != err {
			glog.Error(err)
		}
		object.xCmdObj.Process.Kill()

		object.wg.Wait()

		glog.Info("notify old child exit")
		object.xCmdObj.Close()
		object.xCmdObj = nil
	}

	glog.Infof("wait new child")
	object.xCmdObj = newXCmdObj
	object.wg.Add(1)
	routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
		defer object.wg.Done()

		if err = object.xCmdObj.Wait(); nil != err {
			glog.Error(err)
		}
		if atomic.CompareAndSwapInt32(&object.upgradeFlag, 1, 0) {
			// 正常更新流程
			glog.Infof("child: %d done", object.xCmdObj.Process.Pid)
			return nil
		}

		if 0 == atomic.LoadInt32(&object.killedFlag) {
			// 最大失败重试，直接退出
			object.rebootTimes--
			glog.Errorf("child: %d done unexpected, reboot times countdown: %d",
				object.xCmdObj.Process.Pid,
				object.rebootTimes)
			if 0 > object.rebootTimes {
				os.Exit(-1)
				return nil
			}

			object.xCmdObj.Process.Release()
			object.xCmdObj.Close()
			object.xCmdObj = nil
			object.replaceChildProcess(tcpLnFiles)
		} else {
			glog.Infof("child: %d done", object.xCmdObj.Process.Pid)
		}
		return nil
	}, "DaemonChildWaiter")
	return
}

// notifyChildExit 通知子进程退出
func (object *Daemon) notifyChildExit() (err error) {
	// 通知退出
	object.xCmdObj.ParentWrite([]byte(ExitRequest))
	// 等待回执
	err = object.xCmdObj.ParentRead(func(raw []byte) bool {
		if nil == raw || 0 >= len(raw) {
			return false
		}
		request := string(raw)
		glog.Info("child reply: ", request)
		switch request {
		case ExitReply:
			return false
		default:
			return true
		}
	})
	return
}

// runAsChild 运行于子程序
func (object *Daemon) runAsChild(bootstrapArgs *string,
	logical func(tcpFds map[string]int, ready chan bool), // 业务逻辑
	ready chan bool, // 准备好通道
) {
	// 检查运行参数
	if nil == bootstrapArgs || 0 >= len(*bootstrapArgs) {
		glog.Error("bootstrap argument is empty")
		return
	}

	// 获取通信对象
	object.xCmdObj = xcmd.FromFd(3, 4)
	defer object.xCmdObj.Close()

	// 解析fd
	tcpFds := make(map[string]int)
	panicOnError(json.Unmarshal([]byte(*bootstrapArgs), &tcpFds))

	// 等待业务启动
	routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
		// 等待准备好
		ok := <-ready
		if !ok {
			glog.Error("logical ready not ok")
			object.xCmdObj.ChildWrite([]byte(ReadyError))
			app.GetInstance().Shutdown().WaitComplete()
			return nil
		}

		// 回执启动成功
		glog.Info("child request ready ok")
		object.xCmdObj.ChildWrite([]byte(ReadyOK))
		glog.Info("child request ready ok end")

		// 等待父进程发送退出指令
		object.xCmdObj.ChildRead(func(raw []byte) bool {
			if nil == raw || 0 >= len(raw) {
				glog.Error("parent request nil, shutdown child")
				app.GetInstance().Shutdown().WaitComplete()
				return false
			}
			request := string(raw)
			switch request {
			case ExitRequest:
				glog.Error("parent request exit, shutdown child")
				object.xCmdObj.ChildWrite([]byte(ExitReply))
				app.GetInstance().Shutdown().WaitComplete()
				return false
			}
			return true
		})

		return nil
	}, "Daemon")

	// 将业务逻辑放置在主协程
	logical(tcpFds, ready)
}

// runUpgrade 运行更新
func (object *Daemon) runUpgrade() {
	glog.Info("upgrade app")

	// 读取PID
	raw, err := ioutil.ReadFile(object.pidFile)
	if nil != err {
		glog.Error(err)
		return
	}

	var pid int
	if pid, err = strconv.Atoi(string(raw)); nil != err {
		glog.Error(err)
		return
	}

	// 查找进程
	var p *os.Process
	if p, err = os.FindProcess(pid); nil != err {
		glog.Error(err)
		return
	}

	// 通知更新
	if nil != p {
		if err = p.Signal(syscall.SIGUSR2); nil != err {
			glog.Error(err)
			return
		}
	}
}

// Bootstrap 引导
func (object *Daemon) Bootstrap(tcpPorts map[string]int, // TCP端口
	logical func(tcpFds map[string]int, ready chan bool), // 业务逻辑
	ready chan bool, // 准备好通道
) (err error) {
	rebootTimes := flag.Int("reboot_times", 3, "")
	runInChild := flag.Bool(object.childCmd, false, "run in child")
	runUpgrade := flag.Bool(object.upgradeCmd, false, "run upgrade")
	bootstrapArgs := flag.String(object.bootstrapArgs, "", "bootstrap args")
	flag.Parse()

	// 等待信号
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh)

	// 运行业务逻辑
	if nil != runInChild && *runInChild {
		object.runAsChild(bootstrapArgs, logical, ready)
		return
	}

	// 运行更新程序
	if nil != runUpgrade && *runUpgrade {
		object.runUpgrade()
		return
	}

	// 解析最大重启次数
	if nil != rebootTimes {
		object.rebootTimes = *rebootTimes
	}

	// 保存原始运行参数
	object.origArgs = make([]string, len(os.Args))
	copy(object.origArgs, os.Args)

	// 写进程PID
	panicOnError(ioutil.WriteFile(object.pidFile,
		[]byte(strconv.Itoa(os.Getpid())),
		0666))

	// 清空日志文件
	os.RemoveAll(object.bootstrapLogDir)
	os.Mkdir(object.bootstrapLogDir, 0777)

	// 侦听端口
	tcpLnFiles := make(map[string]*os.File)
	for uniqueName, port := range tcpPorts {
		var ln *net.TCPListener
		ln, err = net.ListenTCP("tcp", &net.TCPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: port,
		})
		if nil != err {
			glog.Error(err)
			return
		}

		var lnFile *os.File
		lnFile, err = ln.File()
		if nil != err {
			glog.Error(err)
			return
		}

		tcpLnFiles[uniqueName] = lnFile
	}

	var ok bool
	ok, err = object.replaceChildProcess(tcpLnFiles)
	if !ok {
		return
	}

	if nil != err {
		glog.Error(err)
		return
	}

	if nil != object.xCmdObj {
		defer object.xCmdObj.Close()
	}

	// 等待信号
parentSignalLoop:
	for s := range signalCh {
		switch s {
		case syscall.SIGINT, syscall.SIGTERM:
			glog.Info("notify child exit")

			// 设置主动停服标志
			atomic.StoreInt32(&object.killedFlag, 1)
			// 通知退出
			if err = object.notifyChildExit(); nil != err {
				glog.Error(err)
			}
			// 发送信号，停止子进程
			if err = object.xCmdObj.Process.Kill(); nil != err {
				glog.Error(err)
			}
			object.wg.Wait()

			break parentSignalLoop

		case syscall.SIGUSR2:
			glog.Infof("notify upgrade app")

			// 设置更新标志
			if !atomic.CompareAndSwapInt32(&object.upgradeFlag, 0, 1) {
				return
			}
			// 替换子进程
			ok, err = object.replaceChildProcess(tcpLnFiles)
			if nil != err {
				glog.Error(err)
			}
			if !ok || nil != err {
				break parentSignalLoop
			}
		}
	}

	glog.Info("daemon exited")
	return
}
