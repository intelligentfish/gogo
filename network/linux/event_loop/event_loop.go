// +build linux

package event_loop

import (
	"github.com/golang/glog"
	"golang.org/x/sys/unix"
	"net"
	"sync"
	"sync/atomic"
)

// 全局变量
var (
	nextEventLoopID int32
)

// Ctx工厂
type CtxFactory func(eventLoop *EventLoop) *Ctx

// closeAllFD 关闭所有FD
func closeAllFD(fds []int) {
	var err error
	for _, fd := range fds {
		if err = unix.Close(fd); nil != err {
			glog.Error(err)
		}
	}
}

// EventLoop 事件循环
type EventLoop struct {
	isMaster       bool              // 是否主节点
	id             int32             // ID
	epFD           int               // Epoll文件描述符
	lnFD           int               // 监听文件描述符
	ctrlRPipe      int               // 控制管道
	ctrlWPipe      int               // 控制管道
	slaveLoops     []*EventLoop      // 从事件循环
	slaveLoopsLock sync.RWMutex      // 从事件循环锁
	epEvents       []unix.EpollEvent // Epoll事件组
	epEventsLock   sync.RWMutex      // Epoll事件组锁
	lbIndex        int               // 负载均衡索引
	totalFD        int32             // 总FD数量
	startFlag      int32             // 启动标志
	stopFlag       int32             // 停止标志
	wg             sync.WaitGroup    // 等待组
	ctxFactory     CtxFactory        // 上下文工厂
	ctxMap         map[int]*Ctx      // 上下文Map
	ctxMapLock     sync.RWMutex      // 上下文Map锁
}

// addCtx 添加上下文
func (object *EventLoop) addCtx(fd int, addr unix.Sockaddr) *Ctx {
	if nil != object.ctxFactory {
		ctx := object.ctxFactory(object)
		if ctx.AcceptEvent(fd, addr) {
			object.ctxMapLock.Lock()
			object.ctxMap[fd] = ctx
			object.ctxMapLock.Unlock()
			glog.Info("loop: ", object.id, ", add ctx to map: ", fd)
			return ctx
		}
		glog.Info("loop: ", object.id, ", delFD: ", fd)
		object.delFD(fd)
	}
	return nil
}

// findCtx 查找上下文
func (object *EventLoop) findCtx(fd int) (ctx *Ctx) {
	glog.Info("loop: ", object.id, ", find ctx")
	object.ctxMapLock.RLock()
	if v, ok := object.ctxMap[fd]; ok {
		ctx = v
	}
	object.ctxMapLock.RUnlock()
	return
}

// delCtx 删除上下文
func (object *EventLoop) delCtx(fd int) (ctx *Ctx) {
	object.ctxMapLock.Lock()
	if v, ok := object.ctxMap[fd]; ok {
		ctx = v
		delete(object.ctxMap, fd)
	}
	object.ctxMapLock.Unlock()
	return
}

// accept 接受
func (object *EventLoop) accept(fd int, addr unix.Sockaddr) *EventLoop {
	ctx := object.addCtx(fd, addr)
	if nil == ctx {
		return object
	}
	err := object.makeFDNonBlock(fd)
	if nil != err {
		glog.Error(err)
		return object
	}
	if err = object.makeFDReadable(int32(fd), true, true); nil != err {
		glog.Error(err)
		return object
	}
	return object
}

// dispatch 派发
func (object *EventLoop) dispatch(fd int, addr unix.Sockaddr) *EventLoop {
	object.slaveLoops[object.lbIndex].accept(fd, addr)
	object.lbIndex = (object.lbIndex + 1) % len(object.slaveLoops)
	return object
}

// addEvent 添加事件
func (object *EventLoop) addEvent(fd int32, events uint32) *EventLoop {
	object.epEventsLock.Lock()
	i := 0
	for ; i < len(object.epEvents); i++ {
		if 0 == object.epEvents[i].Fd {
			object.epEvents[i].Fd = fd
			object.epEvents[i].Events = events
			break
		}
	}
	if i >= len(object.epEvents) {
		t := make([]unix.EpollEvent, 2*cap(object.epEvents))
		copy(t, object.epEvents)
		object.epEvents = t
	}
	object.epEvents[i].Fd = fd
	object.epEvents[i].Events = events
	object.epEventsLock.Unlock()
	atomic.AddInt32(&object.totalFD, 1)
	return object
}

// modEvent 修改事件
func (object *EventLoop) modEvent(fd int32, events uint32) *EventLoop {
	object.epEventsLock.Lock()
	for i := 0; i < len(object.epEvents); i++ {
		if fd == object.epEvents[i].Fd {
			object.epEvents[i].Events = events
			break
		}
	}
	object.epEventsLock.Unlock()
	return object
}

// delEvent 删除事件
func (object *EventLoop) delEvent(fd int32) *EventLoop {
	object.epEventsLock.Lock()
	for i := 0; i < len(object.epEvents); i++ {
		if fd == object.epEvents[i].Fd {
			object.epEvents[i].Fd = 0
			break
		}
	}
	object.epEventsLock.Unlock()
	atomic.AddInt32(&object.totalFD, -1)
	return object
}

// isErrorEvent 是否错误事件
func (object *EventLoop) isErrorEvent(e *unix.EpollEvent) bool {
	return 0 != e.Events&unix.EPOLLERR ||
		0 != e.Events&unix.EPOLLRDHUP
}

// isReadEvent 会否读事件
func (object *EventLoop) isReadEvent(e *unix.EpollEvent) bool {
	return 0 != e.Events&unix.EPOLLIN
}

// isWriteEvent 是否写事件
func (object *EventLoop) isWriteEvent(e *unix.EpollEvent) bool {
	return 0 != e.Events&unix.EPOLLOUT
}

// makeFDNonBlock 使FD非阻塞
func (object *EventLoop) makeFDNonBlock(fd int) (err error) {
	err = unix.SetNonblock(fd, true)
	return
}

// makeSocketReuseAddr 使Socket重用地址
func (object *EventLoop) makeSocketReuseAddr(fd int) (err error) {
	err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	return
}

// makeSocketReusePort 使Socket重用端口
func (object *EventLoop) makeSocketReusePort(fd int) (err error) {
	err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	return
}

// ctlFD 控制FD
func (object *EventLoop) ctrlFD(fd int32,
	addOrMod bool,
	events uint32) (err error) {
	e := unix.EpollEvent{
		Events: events,
		Fd:     fd,
	}
	op := unix.EPOLL_CTL_ADD
	if !addOrMod {
		op = unix.EPOLL_CTL_MOD
	}
	if addOrMod {
		object.addEvent(fd, e.Events)
	} else {
		object.modEvent(fd, e.Events)
	}
	err = unix.EpollCtl(object.epFD, op, int(fd), &e)
	return
}

// makeFDReadable 使FD可读
func (object *EventLoop) makeFDReadable(fd int32, addOrMod, oneShot bool) (err error) {
	events := uint32(unix.EPOLLET | unix.EPOLLIN)
	if oneShot {
		events |= unix.EPOLLONESHOT
	}
	object.ctrlFD(fd, addOrMod, events)
	return
}

// makeFDWriteable 使FD可写
func (object *EventLoop) makeFDWriteable(fd int32, addOrMod, oneShot bool) (err error) {
	events := uint32(unix.EPOLLET | unix.EPOLLOUT)
	if oneShot {
		events |= unix.EPOLLONESHOT
	}
	object.ctrlFD(fd, addOrMod, events)
	return
}

// makeFDReadWriteable 使FD可写
func (object *EventLoop) makeFDReadWriteable(fd int32, addOrMod, oneShot bool) (err error) {
	events := uint32(unix.EPOLLET | unix.EPOLLIN | unix.EPOLLOUT)
	if oneShot {
		events |= unix.EPOLLONESHOT
	}
	object.ctrlFD(fd, addOrMod, events)
	return
}

// delFD 删除FD
func (object *EventLoop) delFD(fd int) (err error) {
	err = unix.EpollCtl(object.epFD, unix.EPOLL_CTL_DEL, fd, nil)
	object.delEvent(int32(fd))
	object.delCtx(fd)
	return
}

// closeAllFD 关闭所有FD
func (object *EventLoop) closeAllFD() {
	object.epEventsLock.Lock()
	defer object.epEventsLock.Unlock()
	var err error
	for i := 0; i < len(object.epEvents); i++ {
		if 0 == object.epEvents[i].Fd {
			continue
		}
		if err = unix.Close(int(object.epEvents[i].Fd)); nil != err {
			glog.Error(err)
		}
	}
}

// loop 循环
func (object *EventLoop) loop() {
	//glog.Info("in master loop? ", object.isMaster)
	defer object.wg.Done()

	var n int
	var err error
loop:
	for 0 == atomic.LoadInt32(&object.stopFlag) {
		if n, err = unix.EpollWait(object.epFD, object.epEvents, -1); nil != err {
			glog.Error(err)
			break loop
		}
		for i := 0; i < n; i++ {
			e := &object.epEvents[i]
			// 退出信号
			if int32(object.ctrlRPipe) == e.Fd {
				glog.Error("notify epoll break loop")
				break loop
			}
			// 错误
			if object.isErrorEvent(e) {
				if int32(object.lnFD) == e.Fd {
					glog.Error("listen fd error: ", e.Events)
					break loop
				}
				glog.Error("error event: ", e.Events)
				object.delFD(int(e.Fd))
				continue
			}
			// 主循环
			if object.isMaster {
				// 主事件循环
				if object.isReadEvent(e) {
					glog.Info("master accept")
					var fd int
					var addr unix.Sockaddr
					for {
						fd, addr, err = unix.Accept(object.lnFD)
						if nil != err {
							if unix.EAGAIN != err {
								glog.Error("unix.Accept error: ", err)
							}
							break
						}
						// 负载均衡到从结点
						//v4Addr := addr.(*unix.SockaddrInet4)
						//glog.Info("dispatch to slave (", fd, ",",
						//	net.IPv4(v4Addr.Addr[0], v4Addr.Addr[1], v4Addr.Addr[2], v4Addr.Addr[3]).String(),
						//	",",
						//	v4Addr.Port,
						//	")")

						object.dispatch(fd, addr)
					}
				}
				continue
			}
			glog.Info("slave loop")
			// 从循环
			if object.isReadEvent(e) {
				glog.Info("slave read")
				ctx := object.findCtx(int(e.Fd))
				glog.Info("findCtx: ", ctx)
				ctx.ReadEvent()
				continue
			}
			if object.isWriteEvent(e) {
				// 可写不做任何处理，外部多协程写
				//	glog.Info("slave write")
				//	object.makeFDWriteable(e.Fd, false, true)
				continue
			}
			glog.Error("unknown event: ", e.Events)
		}
	}
	glog.Info("loop break, master? ", object.isMaster)
	object.Stop()
}

// 工厂方法
func New() (object *EventLoop, err error) {
	object = &EventLoop{}
	if object.epFD, err = unix.EpollCreate1(unix.EPOLL_CLOEXEC); nil != err {
		return
	}
	ctrlPipes := make([]int, 2)
	if err = unix.Pipe(ctrlPipes); nil != err {
		return
	}
	object.id = atomic.AddInt32(&nextEventLoopID, 1)
	object.ctrlRPipe = ctrlPipes[0]
	object.ctrlWPipe = ctrlPipes[1]
	object.slaveLoops = make([]*EventLoop, 0)
	object.epEvents = make([]unix.EpollEvent, 1<<12)
	object.ctxMap = make(map[int]*Ctx)
	return
}

// SetCtxFactory 设置Ctx工厂
func (object *EventLoop) SetCtxFactory(factory CtxFactory) *EventLoop {
	object.ctxFactory = factory
	return object
}

// Listen 侦听
func (object *EventLoop) Listen(port int) (err error) {
	if object.lnFD, err = unix.Socket(unix.AF_INET,
		unix.SOCK_STREAM|unix.O_NONBLOCK,
		unix.IPPROTO_TCP); nil != err {
		return
	}
	if err = object.makeSocketReuseAddr(object.lnFD); nil != err {
		return
	}
	if err = object.makeSocketReusePort(object.lnFD); nil != err {
		return
	}
	addr := &unix.SockaddrInet4{Port: port}
	copy(addr.Addr[:], net.ParseIP("0.0.0.0").To4())
	if err = unix.Bind(object.lnFD, addr); nil != err {
		return
	}
	if err = unix.Listen(object.lnFD, 1<<12); nil != err {
		return
	}
	err = object.makeFDReadable(int32(object.lnFD),
		true,
		false)
	return
}

// Group 组合
func (object *EventLoop) Group(slave *EventLoop) *EventLoop {
	object.slaveLoopsLock.Lock()
	defer object.slaveLoopsLock.Unlock()
	object.isMaster = true
	object.slaveLoops = append(object.slaveLoops, slave)
	return object
}

// Start 启动
func (object *EventLoop) Start() (err error) {
	if !atomic.CompareAndSwapInt32(&object.startFlag, 0, 1) {
		return
	}
	if err = object.makeFDNonBlock(object.ctrlRPipe); nil != err {
		return
	}
	if err = object.makeFDNonBlock(object.ctrlWPipe); nil != err {
		return
	}
	if err = object.makeFDReadable(int32(object.ctrlRPipe),
		true,
		false); nil != err {
		return
	}
	object.wg.Add(1)
	go object.loop()
	return
}

// Stop 停止
func (object *EventLoop) Stop() {
	if !atomic.CompareAndSwapInt32(&object.stopFlag, 0, 1) {
		return
	}
	// 停止监听
	var err error
	if 0 < object.lnFD {
		glog.Info("close ln fd")
		if err = unix.Close(object.lnFD); nil != err {
			glog.Error(err)
		}
	}
	// 终止Epoll循环
	if 0 < object.ctrlWPipe {
		glog.Info("write exit to ctrl pipe")
		if _, err = unix.Write(object.ctrlWPipe, []byte("EXIT")); nil != err {
			glog.Error(err)
		}
		glog.Info("write exit to ctrl done")
	}
	object.wg.Wait()
	// TODO 等待所有处理完成
	if !object.isMaster {
		glog.Info("close all fd")
		object.closeAllFD()
	}
	// 关闭所有文件句柄
	closeAllFD([]int{
		object.ctrlWPipe, /*控制写管道*/
		object.epFD,      /*Epoll文件描述符*/
	})
}
