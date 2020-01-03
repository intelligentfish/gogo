package http_parser

import (
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/byte_buf"
	"strconv"
	"strings"
	"sync"
)

// 部分HTTP头Key
const (
	ContentLengthLowerKey    = "content-length"    // 消息体长度
	ContentTypeLowerKey      = "content-type"      // 消息体类型
	TransferEncodingLowerKey = "transfer-encoding" // 传输编码类型
)

// 协议状态机
type MachineState int

const (
	MachineStateMethod = MachineState(iota)
	MachineStateURI
	MachineStateVersion
	MachineStateHeaderKey
	MachineStateHeaderValue
	MachineStateBody
)

// 解析结果
type ParseResult int

const (
	ParseResultContinue = ParseResult(iota)
	ParseResultOK
	ParseResultError
)

// 解析器
type Parser struct {
	sync.RWMutex
	byteBuf          *byte_buf.ByteBuf
	state            MachineState
	method           string
	uri              string
	version          string
	headers          map[string][]string
	contentLength    int
	contentType      string
	transferEncoding string
	bodyStart        int
	bodyEnd          int
	key              string
	chunkSize        int
}

// 工厂方法s
func New(byteBuf *byte_buf.ByteBuf) *Parser {
	return &Parser{
		byteBuf:   byteBuf,
		state:     MachineStateMethod,
		headers:   make(map[string][]string),
		chunkSize: -1, // 块大小可以为0
	}
}

// withLock 使用锁
func (object *Parser) withLock(read bool, scope func()) {
	if read {
		object.RLock()
	} else {
		object.Lock()
	}
	scope()
	if read {
		object.RUnlock()
	} else {
		object.Unlock()
	}
}

// setMethod 设置方法
func (object *Parser) setMethod(method string) *Parser {
	object.withLock(false, func() {
		object.method = method
	})
	return object
}

// setURI 设置URI
func (object *Parser) setURI(uri string) *Parser {
	object.withLock(false, func() {
		object.uri = uri
	})
	return object
}

// setVersion 设置版本
func (object *Parser) setVersion(version string) *Parser {
	object.withLock(false, func() {
		object.version = version
	})
	return object
}

// addHeader 添加头
func (object *Parser) addHeader(key, value string) *Parser {
	lowerKey := strings.ToLower(key)
	object.withLock(false, func() {
		if v, ok := object.headers[lowerKey]; ok {
			v = append(v, value)
		} else {
			object.headers[lowerKey] = []string{value}
		}
	})
	return object
}

// setContentLength 设置内容长度
func (object *Parser) setContentLength(length int) *Parser {
	object.withLock(false, func() {
		object.contentLength = length
	})
	return object
}

// setContentType 设置内容类型
func (object *Parser) setContentType(contentType string) *Parser {
	object.withLock(false, func() {
		object.contentType = contentType
	})
	return object
}

// setTransferEncoding 设置传输编码
func (object *Parser) setTransferEncoding(value string) *Parser {
	object.withLock(false, func() {
		object.transferEncoding = value
	})
	return object
}

// setBodyStartIndex 设置Body开始索引
func (object *Parser) setBodyStartIndex(index int) *Parser {
	object.withLock(false, func() {
		object.bodyStart = index
	})
	return object
}

// setBodyEndIndex 设置Body结束索引
func (object *Parser) setBodyEndIndex(index int) *Parser {
	object.withLock(false, func() {
		object.bodyEnd = index
	})
	return object
}

// GetMethod 获取方法
func (object *Parser) GetMethod() (method string) {
	object.withLock(true, func() {
		method = object.method
	})
	return
}

// GetURI 获取URI
func (object *Parser) GetURI() (uri string) {
	object.withLock(true, func() {
		uri = object.uri
	})
	return
}

// GetVersion 获取版本
func (object *Parser) GetVersion() (version string) {
	object.withLock(true, func() {
		version = object.version
	})
	return
}

// GetHeaders 获取头
func (object *Parser) GetHeaders(key string) (values []string) {
	lowerKey := strings.ToLower(key)
	object.withLock(true, func() {
		if v, ok := object.headers[lowerKey]; ok && 0 < len(v) {
			values = make([]string, len(v))
			copy(values, v)
		}
	})
	return
}

// GetHeader 获取头
func (object *Parser) GetHeader(key string) (value string) {
	lowerKey := strings.ToLower(key)
	object.withLock(true, func() {
		if v, ok := object.headers[lowerKey]; ok && 0 < len(v) {
			value = v[0]
		}
	})
	return
}

// GetContentLength 获取内容长度
func (object *Parser) GetContentLength() (length int) {
	object.withLock(true, func() {
		length = object.contentLength
	})
	return
}

// GetContentType 获取内容类型
func (object *Parser) GetContentType() (contentType string) {
	object.withLock(true, func() {
		contentType = object.contentType
	})
	return
}

// GetTransferEncoding 获取传输编码
func (object *Parser) GetTransferEncoding() (value string) {
	object.withLock(true, func() {
		value = object.transferEncoding
	})
	return
}

// GetBodyRange 获取消息体范围
func (object *Parser) GetBodyRange() (start, end int) {
	object.withLock(true, func() {
		start, end = object.bodyStart, object.bodyEnd
	})
	return
}

// Reset 重置
func (object *Parser) Reset() *Parser {
	object.state = MachineStateMethod
	object.method = ""
	object.uri = ""
	object.version = ""
	object.headers = make(map[string][]string)
	object.contentLength = 0
	object.contentType = ""
	object.transferEncoding = ""
	object.bodyStart = 0
	object.bodyEnd = 0
	object.key = ""
	object.chunkSize = -1 // 块大小可以为0
	return object
}

// Parse 解析
func (object *Parser) Parse() ParseResult {
	for object.byteBuf.IsReadable() {
		switch object.state {
		case MachineStateMethod:
			// 解析HTTP方法
			method := object.byteBuf.TakeUntil(' ', true)
			if nil != method {
				object.setMethod(string(method))
				object.byteBuf.Skip(' ')
				object.state = MachineStateURI
				continue
			}

			return ParseResultContinue

		case MachineStateURI:
			// 解析HTTP URI
			uri := object.byteBuf.TakeUntil(' ', true)
			if nil != uri {
				object.setURI(string(uri))
				object.byteBuf.Skip(' ')
				object.state = MachineStateVersion
				continue
			}

			return ParseResultContinue

		case MachineStateVersion:
			// 解析HTTP版本
			version := object.byteBuf.TakeUntil('\r', true)
			if nil != version {
				object.setVersion(string(version))
				object.byteBuf.Skip('\r').Skip('\n')
				object.state = MachineStateHeaderKey
				continue
			}

			return ParseResultContinue

		case MachineStateHeaderKey:
			// 解析头Key
			// 流式解析的特点
			if 1 <= object.byteBuf.ReadableBytes() &&
				'\n' == object.byteBuf.PeekByte(object.byteBuf.ReaderIndex()) {
				object.byteBuf.Skip('\n')
			}
			// 流式解析的特点
			if 2 <= object.byteBuf.ReadableBytes() {
				if '\r' == object.byteBuf.PeekByte(object.byteBuf.ReaderIndex()) &&
					'\n' == object.byteBuf.PeekByte(object.byteBuf.ReaderIndex()+1) {
					object.byteBuf.Skip('\r').Skip('\n')
					value := object.GetHeader(ContentTypeLowerKey)
					if "" != value {
						object.setContentType(value)
					}
					value = object.GetHeader(TransferEncodingLowerKey)
					if "" != value {
						object.setTransferEncoding(value)
					}
					object.byteBuf.DiscardReadBytes() //丢弃头
					object.setBodyStartIndex(object.byteBuf.ReaderIndex())
					value = object.GetHeader(ContentLengthLowerKey)
					if "" != value {
						length, err := strconv.Atoi(value)
						if nil != err {
							glog.Error(err)
							return ParseResultError
						}
						object.setContentLength(length)
						object.setBodyEndIndex(object.byteBuf.ReaderIndex() + length)
						object.state = MachineStateBody
						continue
					}

					if "chunked" == object.GetTransferEncoding() {
						object.state = MachineStateBody
						continue
					}

					return ParseResultOK
				}
			}
			keyBytes := object.byteBuf.TakeUntil(':', true)
			if nil != keyBytes {
				object.key = string(keyBytes)
				object.byteBuf.Skip(':')
				object.state = MachineStateHeaderValue
				continue
			}

			return ParseResultContinue

		case MachineStateHeaderValue:
			// 解析头Value
			valueBytes := object.byteBuf.Skip(' ').TakeUntil('\r', true)
			if nil != valueBytes {
				object.addHeader(object.key, string(valueBytes))
				object.byteBuf.Skip('\r').Skip('\n')
				object.state = MachineStateHeaderKey
				continue
			}

			return ParseResultContinue

		case MachineStateBody:
			// 解析消息体
			if "chunked" != object.GetTransferEncoding() {
				//非Chunked 编码
				if 0 >= object.contentLength ||
					object.byteBuf.ReadableBytes() >= object.contentLength {
					return ParseResultOK
				}

				return ParseResultContinue
			} else {
				//Chunked 编码
				if 0 > object.chunkSize {
					readerIndex := object.byteBuf.ReaderIndex()
					chunkSizeBytes := object.byteBuf.TakeUntil('\r', true)
					if nil == chunkSizeBytes {
						chunkSizeBytes = object.byteBuf.TakeUntil(' ', true)
					}
					if nil == chunkSizeBytes {
						return ParseResultContinue
					}

					size, err := strconv.ParseInt(string(chunkSizeBytes), 16, 32)
					if nil != err {
						glog.Error(err)
						return ParseResultError
					}

					object.chunkSize = int(size)
					if 0 == object.chunkSize {
						if 2 == object.byteBuf.ReadableBytes() {
							object.setBodyEndIndex(readerIndex)
							return ParseResultOK
						}

						object.byteBuf.SetReaderIndex(readerIndex)
						return ParseResultContinue
					}

					object.byteBuf.Skip(' ').Skip('\r').Skip('\n').DiscardReadBytes()
				}

				// 流式解析的特点
				if 1 <= object.byteBuf.ReadableBytes() &&
					(' ' == object.byteBuf.PeekByte(object.byteBuf.ReaderIndex()) ||
						'\n' == object.byteBuf.PeekByte(object.byteBuf.ReaderIndex())) {
					object.byteBuf.Skip(' ').Skip('\n')
				}

				if object.chunkSize <= object.byteBuf.ReadableBytes() {
					object.byteBuf.SetReaderIndex(object.byteBuf.ReaderIndex() + object.chunkSize)
					object.chunkSize = -1
					continue
				}

				return ParseResultContinue
			}
		}
	}
	return ParseResultContinue
}
