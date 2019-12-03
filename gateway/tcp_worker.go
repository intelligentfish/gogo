package gateway

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/auto_lock"
	"github.com/intelligentfish/gogo/service"
)

// TCP工作者
type TCPWorker struct {
	auto_lock.AutoLock
	upstream           *Upstream
	sessionLookupTable map[*service.TCPSession]*service.TCPSession
	tcpService         *service.TCPService
}

// 工厂方法
func NewTCPWorker() *TCPWorker {
	object := &TCPWorker{sessionLookupTable: make(map[*service.TCPSession]*service.TCPSession)}
	object.tcpService = service.NewTCPServiceWithCallback(object.onNewSession)
	return object
}

// SetUpstream 设置上游
func (object *TCPWorker) SetUpstream(upstream *Upstream) {
	object.upstream = upstream
}

// onNewSession 新建会话
func (object *TCPWorker) onNewSession(in *service.TCPSession) (blocked bool) {
	in.SetMode(service.TCPSessionModeStream)
	out := service.NewTCPSession().SetMode(service.TCPSessionModeStream)
	out.AddCallback(
		func(session *service.TCPSession, chunk []byte) {
			if !in.IsStopped() {
				in.Write(chunk)
			}
		},
		func(session *service.TCPSession, isRead bool, err error) {
			if isRead {
				in.CloseWrite()
			} else {
				in.CloseRead()
			}
			if out.NeedClose() {
				out.Stop()
				object.WithLock(false, func() {
					delete(object.sessionLookupTable, out)
				})
			}
		})
	addr := fmt.Sprintf("%s:%d", object.upstream.ProxyToHost, object.upstream.ProxyToPort)
	err := out.Connect(addr)
	if nil != err {
		glog.Error(err)
		return true
	}
	in.AddCallback(
		func(session *service.TCPSession, chunk []byte) {
			if !out.IsStopped() {
				out.Write(chunk)
			}
		},
		func(session *service.TCPSession, isRead bool, err error) {
			if isRead {
				out.CloseWrite()
			} else {
				out.CloseRead()
			}
			if in.NeedClose() {
				in.Stop()
				object.WithLock(false, func() {
					delete(object.sessionLookupTable, in)
				})
			}
		})
	object.WithLock(false, func() {
		object.sessionLookupTable[in] = out
		object.sessionLookupTable[out] = in
	})
	out.Start()
	return false
}

// Start 启动方法
func (object *TCPWorker) Start() (err error) {
	err = object.tcpService.StartWithAddr(fmt.Sprintf(":%d", object.upstream.Port))
	return
}
