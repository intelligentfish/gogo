package gateway

import "github.com/intelligentfish/gogo/auto_lock"

// Gateway 网关
type Gateway struct {
	auto_lock.AutoLock
	upstreams []*Upstream
	factory   *Factory
}

// NewGateway 工厂方法
func NewGateway() *Gateway {
	return &Gateway{factory: NewFactory()}
}

// AddUpstream 添加上游
func (object *Gateway) AddUpstream(upstream *Upstream) *Gateway {
	object.WithLock(false, func() {
		object.upstreams = append(object.upstreams, upstream)
	})
	return object
}

// Start 启动
func (object *Gateway) Start() (err error) {
	for _, upstream := range object.upstreams {
		if err = object.factory.Start(upstream); nil != err {
			break
		}
	}
	return
}
