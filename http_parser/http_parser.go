package http_parser

// HTTP协议解析器状态
type HTTPParserStatus int

const (
	HTTPParserStatusContinue = HTTPParserStatus(iota) // 继续
	HTTPParserStatusError                             // 失败
	HTTPParserStatusOK                                // 成功
)

type ValueHook func(value []byte)
type KeyValueHook func(key, value []byte)

// HTTPParser HTTP协议解析器
type HTTPParser interface {
	// 重置
	Reset()
	// 处理
	Process(buffer []byte) HTTPParserStatus
	// 处理方法
	SetMethodHook(hook ValueHook) HTTPParser
	// 处理URL
	SetURLHook(hook ValueHook) HTTPParser
	// 处理协议
	SetProtocolHook(hook ValueHook) HTTPParser
	// 处理头
	SetHeaderHook(hook KeyValueHook) HTTPParser
	// 处理消息体
	SetBodyHook(hook ValueHook) HTTPParser
	// 处理块
	SetChunkedHook(hook ValueHook) HTTPParser
}
