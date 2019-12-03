package gateway

import "github.com/intelligentfish/gogo/auto_lock"

// Factory 工厂
type Factory struct {
	auto_lock.AutoLock
	workerMap map[UpstreamType]Worker
}

// NewFactory 工厂方法
func NewFactory() *Factory {
	return &Factory{
		workerMap: map[UpstreamType]Worker{
			UpstreamTypeTCP:  NewTCPWorker(),
			UpstreamTypeHTTP: NewHTTPWorker(),
		},
	}
}
// 工厂方法
func (object *Factory) makeWorker(upstream *Upstream) (worker Worker) {
	object.WithLock(true, func() {
		worker = object.workerMap[upstream.UpstreamType]
	})
	return
}

// Start 启动
func (object *Factory) Start(upstream *Upstream) (err error) {
	worker := object.makeWorker(upstream)
	worker.SetUpstream(upstream)
	return worker.Start()
}
