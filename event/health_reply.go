package event

// HealthCheckReply 健康检查响应
type HealthCheckReply struct {
	UniqueName string // 名称
	OK         bool   // 成功
}
