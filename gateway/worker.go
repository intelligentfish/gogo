package gateway

// 工作者
type Worker interface {
	Start() error
	SetUpstream(upstream *Upstream)
}
