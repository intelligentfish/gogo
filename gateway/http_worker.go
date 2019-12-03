package gateway

import (
	"fmt"
	"github.com/intelligentfish/gogo/service"
)

// HTTPWorker HTTP工作者
type HTTPWorker struct {
	tcpService *service.TCPService
	upstream   *Upstream
}

// 工厂方法
func NewHTTPWorker() *HTTPWorker {
	return &HTTPWorker{tcpService: service.NewTCPService()}
}

// SetUpstream 设置上游
func (object *HTTPWorker) SetUpstream(upstream *Upstream) {
	object.upstream = upstream
}

// Start 启动
func (object *HTTPWorker) Start() (err error) {
	addr := fmt.Sprintf(":%d", object.upstream.Port)
	return object.tcpService.StartWithAddr(addr)
}
