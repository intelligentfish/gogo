// +build !windows

package main

import (
	"fmt"
	"github.com/intelligentfish/gogo/daemon"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func main() {
	ready := make(chan bool, 1)
	daemon := daemon.Default()
	daemon.Bootstrap(map[string]int{
		"web": 10080,
	}, func(tcpFds map[string]int, exitCh <-chan interface{}) {
		engine := gin.Default()
		engine.GET("/pid", func(ctx *gin.Context) {
			ctx.String(http.StatusOK, fmt.Sprintf("pid:%d\n", os.Getpid()))
		})

		// 模拟耗时的操作...
		time.Sleep(10 * time.Second)

		fd := tcpFds["web"]
		f := os.NewFile(uintptr(fd), "web")
		listener, err := net.FileListener(f)
		if err != nil {
			ready <- false
			glog.Error(err)
			return
		}

		defer listener.Close()
		ready <- true
		go func() {
			<-exitCh
			listener.Close()
		}()
		if err = engine.RunListener(listener); nil != err {
			glog.Error(err)
		}
	}, ready)
}
