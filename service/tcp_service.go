package service

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/app_cfg"
	"github.com/intelligentfish/gogo/auto_lock"
	"github.com/intelligentfish/gogo/byte_buffer_pool"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/fix_slice_pool"
	"github.com/intelligentfish/gogo/priority_define"
	"github.com/intelligentfish/gogo/routine_pool"
	"io"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
)

// 常量
const (
	defaultWriteChSize = 16 // 默认写通道大小
)

// 变量
var (
	nextSessionId          int32           // 下一个会话id
	tcpSessionPoolOnce     sync.Once       // 执行一次
	tcpSessionPoolInstance *TCPSessionPool // 会话池
)

// TCPSessionPool TCP会话池
type TCPSessionPool struct {
	pool *sync.Pool
}

// Borrow 获取TCP会话
func (object *TCPSessionPool) Borrow() *TCPSession {
	return object.pool.Get().(*TCPSession)
}

// Return 退还TCP会话
func (object *TCPSessionPool) Return(tcpSession *TCPSession) {
	object.pool.Put(tcpSession)
}

// GetTCPSessionPoolInstance 获取单实例
func GetTCPSessionPoolInstance() *TCPSessionPool {
	tcpSessionPoolOnce.Do(func() {
		tcpSessionPoolInstance = &TCPSessionPool{
			pool: &sync.Pool{
				New: func() interface{} {
					return &TCPSession{}
				},
			},
		}
	})
	return tcpSessionPoolInstance
}

// TCP会话模式
type TCPSessionMode int

const (
	TCPSessionModeChunk  = TCPSessionMode(iota) // 块模式
	TCPSessionModeStream                        // 流模式
)

// TCPSession TCP会话
type TCPSession struct {
	auto_lock.AutoLock
	ID                int            // 会话id
	Mode              TCPSessionMode // 模式
	C                 net.Conn       // TCP连接
	dataCallbackList  []DataCallback // 数据回调
	errorCallbackList []ErrCallback  // 错误回调
	writeCh           chan []byte    // 写通道
	stoppedReadFlag   int32
	stoppedWriteFlag  int32
	stopFlag          int32          // 停止标志
	readWG            sync.WaitGroup // 等待组
	writeWG           sync.WaitGroup // 等待组
}

// 数据回调
type DataCallback func(session *TCPSession, chunk []byte)

// 错误回调
type ErrCallback func(session *TCPSession, isRead bool, err error)

// NewTCPSession 工厂方法
func NewTCPSession() *TCPSession {
	object := GetTCPSessionPoolInstance().Borrow()
	object.ID = int(atomic.AddInt32(&nextSessionId, 1))
	object.writeCh = make(chan []byte, defaultWriteChSize)
	object.stopFlag = 0
	return object
}

// 设置模式
func (object *TCPSession) SetMode(mode TCPSessionMode) *TCPSession {
	object.Mode = mode
	return object
}

// SetConn 设置连接
func (object *TCPSession) SetConn(c net.Conn) *TCPSession {
	object.C = c
	return object
}

// AddCallback 添加回调
func (object *TCPSession) AddCallback(dataCallback DataCallback, errorCallback ErrCallback) *TCPSession {
	object.WithLock(false,
		func() {
			object.dataCallbackList = append(object.dataCallbackList, dataCallback)
			object.errorCallbackList = append(object.errorCallbackList, errorCallback)
		})
	return object
}

// Connect 连接
func (object *TCPSession) Connect(addr string) (err error) {
	object.C, err = net.Dial("tcp", addr)
	if nil != err {
		return
	}
	return
}

// 关闭读
func (object *TCPSession) CloseRead() {
	atomic.StoreInt32(&object.stoppedReadFlag, 1)
	object.C.(*net.TCPConn).CloseRead()
}

// 关闭写
func (object *TCPSession) CloseWrite() {
	atomic.StoreInt32(&object.stoppedWriteFlag, 1)
	object.C.(*net.TCPConn).CloseWrite()
}

// 需要关闭
func (object *TCPSession) NeedClose() bool {
	return 1 == atomic.LoadInt32(&object.stoppedReadFlag) &&
		1 == atomic.LoadInt32(&object.stoppedWriteFlag)
}

// 读
func (object *TCPSession) read() {
	var n, bodySize int
	var err error
	chunk := fix_slice_pool.GetFixSlicePoolInstance().BorrowSlice(8192)
	cache := byte_buffer_pool.GetByteBufferPoolInstance().BorrowByteBuffer()
	for {
		n, err = object.C.Read(chunk)
		if nil != err {
			break
		}
		//fmt.Println("read:\n", string(chunk[:n]))
		if 0 >= n {
			err = io.EOF
			break
		}
		switch object.Mode {
		// 块模式
		case TCPSessionModeChunk:
			cache.Write(chunk[:n])
			for 4 <= cache.Len() {
				if 0 == bodySize {
					bodySize = int(binary.BigEndian.Uint32(cache.Next(4)))
				}
				if bodySize > cache.Len() {
					break
				}
				data := cache.Next(bodySize)
				object.WithLock(false,
					func() {
						for _, callback := range object.dataCallbackList {
							callback(object, data)
						}
					})
				bodySize = 0
			}
		case TCPSessionModeStream:
			// 流模式
			object.WithLock(false,
				func() {
					for _, callback := range object.dataCallbackList {
						callback(object, chunk[:n])
					}
				})
		}
	}
	object.readWG.Done()
	if nil != err && !object.IsStopped() {
		glog.Errorf("session: %d error: %s", object.ID, err)
		object.WithLock(true, func() {
			for _, callback := range object.errorCallbackList {
				callback(object, true, err)
			}
		})
	}
	fix_slice_pool.GetFixSlicePoolInstance().ReturnSlice(len(chunk), chunk)
	byte_buffer_pool.GetByteBufferPoolInstance().ReturnByteBuffer(cache)
}

// 写空
func (object *TCPSession) writeUntilEmpty(raw []byte) (err error) {
	var n int
	writeSize := 0
	bodySize := len(raw)
	for bodySize > writeSize {
		n, err = object.C.Write(raw)
		if nil != err {
			break
		}
		if 0 >= n {
			err = io.EOF
			break
		}
		writeSize += n
	}
	return
}

// 写
func (object *TCPSession) write() {
	var err error
	header := fix_slice_pool.GetFixSlicePoolInstance().BorrowSlice(4)
	for {
		body, ok := <-object.writeCh
		if !ok || nil == body {
			break
		}
		//fmt.Println("write:\n", string(body))
		switch object.Mode {
		// 块模式
		case TCPSessionModeChunk:
			binary.BigEndian.PutUint32(header, uint32(len(body)))
			if err = object.writeUntilEmpty(header); nil != err {
				break
			}
		}
		if err = object.writeUntilEmpty(body); nil != err {
			break
		}
	}
	if nil != err && !object.IsStopped() {
		glog.Errorf("session: %d error: %s", object.ID, err)
		object.WithLock(true,
			func() {
				for _, callback := range object.errorCallbackList {
					callback(object, false, err)
				}
			})
	}
	object.writeWG.Done()
	fix_slice_pool.GetFixSlicePoolInstance().ReturnSlice(4, header)
}

// IsStopped 是否已停止
func (object *TCPSession) IsStopped() bool {
	return 1 == atomic.LoadInt32(&object.stopFlag)
}

// Start 启动
func (object *TCPSession) Start() {
	object.readWG.Add(1)
	object.writeWG.Add(1)
	routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
		object.read()
		return nil
	}, fmt.Sprintf(`TCPSession-%d Reader`, object.ID))
	routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
		object.write()
		return nil
	}, fmt.Sprintf(`TCPSession-%d Writer`, object.ID))
}

// Write 写
func (object *TCPSession) Write(raw []byte) {
	object.writeCh <- raw
}

// Stop 停止
func (object *TCPSession) Stop() {
	if !atomic.CompareAndSwapInt32(&object.stopFlag, 0, 1) {
		return
	}

	object.writeCh <- nil
	object.writeWG.Wait()
	object.C.Close()
	object.readWG.Wait()
	close(object.writeCh)

	GetTCPSessionPoolInstance().Return(object)
}

// TCPService TCP服务
type TCPService struct {
	ln                 net.Listener
	stopFlag           int32
	newSessionCallback NewSessionCallback
}

// 新建会话回调
type NewSessionCallback func(session *TCPSession) (blocked bool)

// 工厂方法
func NewTCPService() *TCPService {
	object := &TCPService{}
	event_bus.GetInstance().MountingOnce(reflect.TypeOf(&event.AppShutdownEvent{}),
		"TCPService",
		func(param interface{}) {
			if priority_define.TCPServiceShutdownPriority !=
				param.(*event.AppShutdownEvent).ShutdownPriority {
				return
			}

			object.Stop()
			glog.Info("TCPService done")
		})
	return object
}

// 工厂方法
func NewTCPServiceWithCallback(newSessionCallback NewSessionCallback) *TCPService {
	object := NewTCPService()
	object.newSessionCallback = newSessionCallback
	return object
}

// handleConnection 处理连接
func (object *TCPService) handleConnection(c net.Conn) {
	session := NewTCPSession().SetConn(c).AddCallback(
		func(session *TCPSession, chunk []byte) {
			if !session.IsStopped() {
				//Debug...
			}
		},
		func(session *TCPSession, isRead bool, err error) {
			//Debug
		})
	if nil != object.newSessionCallback {
		if object.newSessionCallback(session) {
			session.Stop()
			return
		}
	}
	session.Start()
}

// Start 启动
func (object *TCPService) Start() (err error) {
	addr := fmt.Sprintf(`:%d`, app_cfg.GetInstance().TCPServiceConfig.Port)
	return object.StartWithAddr(addr)
}

// StartWithAddr 启动
func (object *TCPService) StartWithAddr(addr string) (err error) {
	object.ln, err = net.Listen("tcp", addr)
	if nil != err {
		return
	}

	routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
		var c net.Conn
		var err error
		for {
			c, err = object.ln.Accept()
			if object.IsStopped() {
				break
			}

			if nil != err {
				glog.Error(err)
				continue
			}

			object.handleConnection(c)
		}
		return nil
	}, "TCPService")
	return
}

// IsStopped 是否已停止
func (object *TCPService) IsStopped() bool {
	return 1 == atomic.LoadInt32(&object.stopFlag)
}

// Stop 停止
func (object *TCPService) Stop() {
	if !atomic.CompareAndSwapInt32(&object.stopFlag, 0, 1) {
		return
	}

	object.ln.Close()
}
