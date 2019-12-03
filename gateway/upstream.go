package gateway

// UpstreamType 上游类型
type UpstreamType int

const (
	UpstreamTypeTCP       = UpstreamType(iota) // TCP
	UpstreamTypeUDP                            // UDP
	UpstreamTypeHTTP                           // HTTP
	UpstreamTypeHTTTPS                         // HTTPS
	UpstreamTypeWebSocket                      // WebSocket
)

// Upstream 上游
type Upstream struct {
	UpstreamType UpstreamType // 上游类型
	Port         int          // 端口
	ProxyToHost  string       // 代理主机
	ProxyToPort  int          // 代理端口
	URIs         []string     // URI
	ConnUUIDs    []string     // 连接UUID(公网到内网反向注册代理)
}

// 工厂方法
func NewUpstream(upstreamType UpstreamType,
	port int,
	proxyHost string,
	proxyPort int,
	uris []string,
	connUUIDs []string) *Upstream {
	return &Upstream{
		UpstreamType: upstreamType,
		Port:         port,
		ProxyToHost:  proxyHost,
		ProxyToPort:  proxyPort,
		URIs:         uris,
		ConnUUIDs:    connUUIDs,
	}
}
