package event_loop

import (
	"context"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/byte_buf"
	"github.com/intelligentfish/gogo/routine_pool"
	"golang.org/x/sys/unix"
	"net"
	"sync/atomic"
)

// 接受事件钩子
type AcceptEventHook func() bool

// 读事件钩子
type ReadEventHook func(buf *byte_buf.ByteBuf, err error)

// 写事件钩子
type WriteEventHook func(buf *byte_buf.ByteBuf, err error)

// Ctx 处理器
type Ctx struct {
	eventLoop       *EventLoop
	fd              int
	addr            *unix.SockaddrInet4
	v4ip            net.IP
	port            int
	readEndFlag     int32
	writeEndFlag    int32
	acceptEventHook AcceptEventHook
	readEventHook   ReadEventHook
	writeEventHook  WriteEventHook
}

// 上下文选项
type CtxOption func(ctx *Ctx)

// EventLoopOption EventLoop选项
func EventLoopOption(eventLoop *EventLoop) CtxOption {
	return func(ctx *Ctx) {
		ctx.eventLoop = eventLoop
	}
}

// AcceptEventHookOption 接受事件钩子选项
func AcceptEventHookOption(hook AcceptEventHook) CtxOption {
	return func(ctx *Ctx) {
		ctx.acceptEventHook = hook
	}
}

// ReadEventHookOption 读取事件钩子选项
func ReadEventHookOption(hook ReadEventHook) CtxOption {
	return func(ctx *Ctx) {
		ctx.readEventHook = hook
	}
}

// WriteEventHookOption 写事件钩子选项
func WriteEventHookOption(hook WriteEventHook) CtxOption {
	return func(ctx *Ctx) {
		ctx.writeEventHook = hook
	}
}

// shutdownSocket 关闭socket
func (object *Ctx) shutdownSocket(read bool) {
	//glog.Info("shutdownSocket: ", read, ",", object.readEndFlag, ",", object.writeEndFlag)
	if read && atomic.CompareAndSwapInt32(&object.readEndFlag, 0, 1) {
		unix.Shutdown(object.fd, unix.SHUT_RD)
	}
	if !read && atomic.CompareAndSwapInt32(&object.writeEndFlag, 0, 1) {
		unix.Shutdown(object.fd, unix.SHUT_WR)
	}
	if 1 == atomic.LoadInt32(&object.readEndFlag) &&
		1 == atomic.LoadInt32(&object.writeEndFlag) {
		object.Close()
	}
}

// NewCtx 工厂方法
func NewCtx(options ...CtxOption) *Ctx {
	object := &Ctx{
		readEndFlag:  1,
		writeEndFlag: 1,
	}
	for _, option := range options {
		option(object)
	}
	return object
}

// SetOption 设置可选项
func (object *Ctx) SetOption(options ...CtxOption) *Ctx {
	for _, option := range options {
		option(object)
	}
	return object
}

// GetV4IP 获取IP地址
func (object *Ctx) GetV4IP() string {
	return object.v4ip.String()
}

// GetPort 获取端口
func (object *Ctx) GetPort() int {
	return object.addr.Port
}

// Close 关闭
func (object *Ctx) Close() {
	glog.Infof("Close: (%d,%s:%d)",
		object.eventLoop.id,
		object.GetV4IP(),
		object.GetPort())

	object.eventLoop.delFD(object.fd)
}

// AcceptEvent 接受
func (object *Ctx) AcceptEvent(fd int, addr unix.Sockaddr) bool {
	object.fd = fd
	object.addr = addr.(*unix.SockaddrInet4)
	object.v4ip = net.IPv4(object.addr.Addr[0],
		object.addr.Addr[1],
		object.addr.Addr[2],
		object.addr.Addr[3]).To4()

	glog.Infof("AcceptEvent: (%d,%s:%d)",
		object.eventLoop.id,
		object.GetV4IP(),
		object.GetPort())

	if nil != object.acceptEventHook {
		return object.acceptEventHook()
	}
	return true
}

// ReadEvent 处理读
func (object *Ctx) ReadEvent() {
	atomic.StoreInt32(&object.readEndFlag, 0)
	routine_pool.GetInstance().CommitTask(func(ctx context.Context, params []interface{}) {
		var n int
		var err error
		buf := byte_buf.GetPoolInstance().Borrow(byte_buf.InitCapOption(1 << 13))
		for {
			atomic.StoreInt32(&object.readEndFlag, 0)
			n, err = unix.Read(object.fd, buf.Internal()[buf.WriterIndex():buf.InitCap()])
			atomic.StoreInt32(&object.readEndFlag, 1)
			if nil != err {
				if unix.EAGAIN == err {
					err = nil
					object.eventLoop.makeFDReadable(int32(object.fd),
						false,
						true)
				}
				break
			}
			if 0 == n {
				//读关闭
				object.shutdownSocket(true)
				break
			}
			buf.SetWriterIndex(buf.WriterIndex() + n)
		}
		object.readEventHook(buf, err)
	}, "CtxRead")
}

// WriteEvent 处理写
func (object *Ctx) WriteEvent(buf *byte_buf.ByteBuf, err error) {
	object.writeEventHook(buf, err)
}

// Write 异步写
func (object *Ctx) Write(buf *byte_buf.ByteBuf) {
	atomic.StoreInt32(&object.writeEndFlag, 0)
	routine_pool.GetInstance().CommitTask(func(ctx context.Context, params []interface{}) {
		var n int
		var err error
		for buf.IsReadable() {
			atomic.StoreInt32(&object.writeEndFlag, 0)
			n, err = unix.Write(object.fd, buf.Internal()[buf.ReaderIndex():buf.WriterIndex()])
			atomic.StoreInt32(&object.writeEndFlag, 1)
			if nil != err {
				if unix.EAGAIN == err {
					err = nil
					object.eventLoop.makeFDWriteable(int32(object.fd),
						false,
						true)
				}
				break
			}
			if 0 == n {
				//写关闭
				object.shutdownSocket(false)
				break
			}
			buf.SetReaderIndex(buf.ReaderIndex() + n)
		}
		object.writeEventHook(buf, nil)
	}, "CtxWrite")
}
