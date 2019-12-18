package priority_define

// 关闭优先级
type ShutdownPriority int

const (
	ShutdownPriorityUnknown          = ShutdownPriority(iota)
	DBShutdownPriority               // 数据库关闭优先级
	ServiceRegisterDiscoveryPriority // 服务注册发现
	RSAGeneratorShutdownPriority     // RSA生成器关闭优先级
	AssetsShutdownPriority           // 资产关闭优先级
	MQShutdownPriority               // MQ关闭优先级
	RPCServiceShutdownPriority       // RPC服务优先级
	HTTPServiceShutdownPriority      // HTTP服务优先级
	TCPServiceShutdownPriority       // TCP服务优先级
	ConfigClientShutdownPriority     // 配置客户端优先级
	HealthCheckerPriority            // 健康检查器优先级
	ShutdownPriorityMax              // 最高优先级
)
