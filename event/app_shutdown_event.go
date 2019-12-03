package event

import "github.com/intelligentfish/gogo/priority_define"

// AppShutdownEvent 应用关闭事件
type AppShutdownEvent struct {
	priority_define.ShutdownPriority // 关闭优先级
}
