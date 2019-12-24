package service

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/app_cfg"
	"github.com/intelligentfish/gogo/auto_lock"
	"github.com/intelligentfish/gogo/buffer"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/fix_slice_pool"
	"github.com/intelligentfish/gogo/priority_define"
	"github.com/intelligentfish/gogo/routine_pool"
	"github.com/intelligentfish/gogo/spin_lock"
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
	debug              bool               // 调试标志
	name               string             // 会话名
	ID                 int                // 会话id
	newRoutineSpinLock spin_lock.SpinLock // 自旋锁
	Mode               TCPSessionMode     // 模式
	C                  net.Conn           // TCP连接
	dataCallbackList   []DataCallback     // 数据回调
	errorCallbackList  []ErrCallback      // 错误回调
	writeCh            chan []byte        // 写通道
	writeChSizeLock    sync.Mutex         // 写通道长度锁
	stoppedReadFlag    int32              // 停止读标志
	stoppedWriteFlag   int32              // 停止写标志
	stopFlag           int32              // 停止标志
	readWG             sync.WaitGroup     // 等待组
	writeWG            sync.WaitGroup     // 等待组
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

// 设置调试标志
func (object *TCPSession) SetDebug() *TCPSession {
	object.debug = true
	return object
}

// 设置名称
func (object *TCPSession) SetName(name string) *TCPSession {
	object.name = name
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
	var n int
	var err error
	readBuf := buffer.GetPoolInstance().Borrow(1 << 13)
	for {
		n, err = object.C.Read(readBuf.Internal[readBuf.GetWriteIndex():])
		if nil != err {
			break
		}
		if object.debug {
			fmt.Printf("op read (%s-%d, %d, %v)\n", object.name, object.ID, n, err)
		}
		if 0 >= n {
			err = io.EOF
			break
		}
		switch object.Mode {
		// 块模式
		case TCPSessionModeChunk:
			readBuf.SetWriteIndex(readBuf.GetWriteIndex() + n)
			for 4 <= readBuf.ReadableBytes() {
				chunkSize := int(binary.BigEndian.Uint32(readBuf.Slice(4)))
				if chunkSize+4 > readBuf.ReadableBytes() {
					break
				}
				readBuf.SetReadIndex(readBuf.GetReadIndex() + 4)
				object.WithLock(false,
					func() {
						for _, callback := range object.dataCallbackList {
							callback(object, readBuf.Slice(chunkSize))
						}
					})
				readBuf.SetReadIndex(readBuf.GetReadIndex() + chunkSize)
				readBuf.DiscardReadBytes()
			}
		case TCPSessionModeStream:
			// 流模式
			object.WithLock(false,
				func() {
					for _, callback := range object.dataCallbackList {
						callback(object, readBuf.Slice(n))
					}
				})
			readBuf.SetReadIndex(readBuf.GetWriteIndex()).DiscardReadBytes()
		}
	}
	object.readWG.Done()
	if nil != err && !object.IsStopped() {
		if object.debug {
			glog.Errorf("session: %d error: %s", object.ID, err)
		}
		object.WithLock(true, func() {
			for _, callback := range object.errorCallbackList {
				callback(object, true, err)
			}
		})
	}
	buffer.GetPoolInstance().Return(readBuf.SetReadIndex(readBuf.GetWriteIndex()).DiscardReadBytes())
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
	if object.debug {
		fmt.Printf("op write (%s-%d, %d, %v)\n", object.name, object.ID, n, err)
	}
	return
}

// 写
func (object *TCPSession) write() {
	object.writeWG.Add(1)
	defer object.writeWG.Done()

	var err error
	needClosed := false
	continueFlag := true
	header := fix_slice_pool.GetFixSlicePoolInstance().BorrowSlice(4)
	for continueFlag {
		object.writeChSizeLock.Lock()
		continueFlag = 0 != len(object.writeCh)
		object.writeChSizeLock.Unlock()
		if !continueFlag {
			break
		}

		body, ok := <-object.writeCh
		if !ok || nil == body {
			needClosed = true
			break
		}

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
	if needClosed || nil != err {
		if !object.IsStopped() {
			if object.debug {
				glog.Errorf("session: %d error: %s", object.ID, err)
			}
			object.WithLock(true,
				func() {
					for _, callback := range object.errorCallbackList {
						callback(object, false, err)
					}
				})
		}
	}
	fix_slice_pool.GetFixSlicePoolInstance().ReturnSlice(4, header)
}

// IsStopped 是否已停止
func (object *TCPSession) IsStopped() bool {
	return 1 == atomic.LoadInt32(&object.stopFlag)
}

// Start 启动
func (object *TCPSession) Start() {
	object.readWG.Add(1)
	routine_pool.GetInstance().CommitTask(func(ctx context.Context, params []interface{}) {
		object.read()
		return
	}, fmt.Sprintf(`TCPSession-%d Reader`, object.ID))
}

// Write 写
func (object *TCPSession) Write(raw []byte) {
	object.writeCh <- raw

	object.writeChSizeLock.Lock()
	chSize := len(object.writeCh)
	object.writeChSizeLock.Unlock()
	if 1 == chSize {
		// 自旋
		object.newRoutineSpinLock.Lock()

		// 开启协程写操作
		routine_pool.GetInstance().CommitTask(func(ctx context.Context, params []interface{}) {
			defer object.newRoutineSpinLock.Unlock()
			object.write()
			return
		}, fmt.Sprintf(`TCPSession-%d Writer`, object.ID))
	}
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

	object.newRoutineSpinLock.Unlock()
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
		func(ctx context.Context, param interface{}) {
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

	routine_pool.GetInstance().CommitTask(func(ctx context.Context, params []interface{}) {
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
		return
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
