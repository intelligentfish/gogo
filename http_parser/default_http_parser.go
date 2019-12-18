package http_parser

// MachineState HTTP 状态机
type MachineState int

const (
	MachineStateMethod = MachineState(iota)
	MachineStateURL
	MachineStateProtocol
	MachineStateHeaderKey
	MachineStateHeaderValue
	MachineStateCRLF
	MachineStateBody
)

var (
	contentLengthKey = "content-length"
)

// 方法解析器
type methodParser struct {
	start int
}

// 重置
func (object *methodParser) Reset() *methodParser {
	object.start = 0
	return object
}

// 解析
func (object *methodParser) Parse(buf []byte) (result HTTPParserStatus, begin, end int) {
	for object.start < len(buf) && ' ' == buf[object.start] {
		object.start++
	}
	start := object.start
	for object.start < len(buf) && ' ' != buf[object.start] {
		object.start++
	}
	if object.start >= len(buf) {
		return HTTPParserStatusContinue, -1, -1
	}
	return HTTPParserStatusOK, start, object.start
}

// URL 解析器
type urlParser struct {
	methodParser
}

// 协议解析器
type protocolParser struct {
	methodParser
}

// 请求头Key解析器
type headerKeyParser struct {
	start int
}

// 重置
func (object *headerKeyParser) Reset() *headerKeyParser {
	object.start = 0
	return object
}

// 解析
func (object *headerKeyParser) Parse(buf []byte) (result HTTPParserStatus, begin, end int) {
	for object.start < len(buf) && ' ' == buf[object.start] {
		object.start++
	}
	start := object.start
	for object.start < len(buf) && ':' != buf[object.start] {
		object.start++
	}
	if object.start >= len(buf) {
		return HTTPParserStatusContinue, -1, -1
	}
	return HTTPParserStatusOK, start, object.start
}

// DefaultHTTPParser 默认HTTP解析器
type DefaultHTTPParser struct {
	start           int
	contentLength   int
	machineState    MachineState
	methodHook      ValueHook
	urlHook         ValueHook
	protocolHook    ValueHook
	bodyHook        ValueHook
	chunkedHook     ValueHook
	headerHook      KeyValueHook
	methodParser    methodParser
	urlParser       urlParser
	protocolParser  protocolParser
	headerKeyParser headerKeyParser
}

// 工厂方法
func NewDefaultHTTPParser() HTTPParser {
	return &DefaultHTTPParser{}
}

// Reset 重置
func (object *DefaultHTTPParser) Reset() {
	object.start = 0
	object.contentLength = 0
	object.machineState = MachineStateMethod
}

// Process 处理
func (object *DefaultHTTPParser) Process(buf []byte) HTTPParserStatus {
	//TODO
	for {
		switch object.machineState {
		case MachineStateMethod:
			// 解析HTTP方法
			result, begin, end := object.methodParser.Parse(buf[object.start:])
			if HTTPParserStatusOK != result {
				return result
			}
			if nil != object.methodHook {
				object.methodHook(buf[object.start+begin : object.start+end])
			}
			object.start = end + 1
			object.machineState = MachineStateURL

		case MachineStateURL:
			// 解析HTTP URL
			result, begin, end := object.urlParser.Parse(buf[object.start:])
			if HTTPParserStatusOK != result {
				return result
			}
			if nil != object.urlHook {
				object.urlHook(buf[object.start+begin : object.start+end])
			}
			object.start = end + 1

		case MachineStateProtocol:
			// 解析HTTP协议
			result, begin, end := object.protocolParser.Parse(buf[object.start:])
			if HTTPParserStatusOK != result {
				return result
			}
			if nil != object.protocolHook {
				object.protocolHook(buf[object.start+begin : object.start+end])
			}
			object.start = end + 1

		case MachineStateCRLF:
			// 解析CRLF
		}
	}
	return HTTPParserStatusContinue
}

// SetMethodHook 处理方法
func (object *DefaultHTTPParser) SetMethodHook(hook ValueHook) HTTPParser {
	object.methodHook = hook
	return object
}

// SetURLHook 处理URL
func (object *DefaultHTTPParser) SetURLHook(hook ValueHook) HTTPParser {
	object.urlHook = hook
	return object
}

// SetProtocolHook 处理协议
func (object *DefaultHTTPParser) SetProtocolHook(hook ValueHook) HTTPParser {
	object.protocolHook = hook
	return object
}

// SetHeaderHook 处理头
func (object *DefaultHTTPParser) SetHeaderHook(hook KeyValueHook) HTTPParser {
	object.headerHook = hook
	return object
}

// SetBodyHook 处理消息体
func (object *DefaultHTTPParser) SetBodyHook(hook ValueHook) HTTPParser {
	object.bodyHook = hook
	return object
}

// SetChunkedHook 处理块
func (object *DefaultHTTPParser) SetChunkedHook(hook ValueHook) HTTPParser {
	object.chunkedHook = hook
	return object
}
