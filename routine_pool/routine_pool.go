package routine_pool

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/app"
	"github.com/intelligentfish/gogo/app_cfg"
	"math"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

// 常量
const (
	defaultTaskQueueSize = 1 << 20 // 默认任务队列长度
)

// 变量
var (
	routinePoolOnce      sync.Once                              // 执行一次
	routinePoolInstance  *RoutinePool                           // 协程池
	nextRunnableID       = int64(0)                             // 下一个执行体id
	ErrRoutinePoolClosed = errors.New("RoutinePool was closed") // 协程池已关闭错误
)

// Runnable 执行体
type Runnable interface {
	GetID() string                        // ID
	GetParams() []interface{}             // 获取参数
	GetCtx() context.Context              // 获取上下文
	Run(params []interface{}) interface{} // 执行
}

// DefaultRunnable 默认实现
type DefaultRunnable struct {
	ID     int64
	Name   string
	Method func(ctx context.Context, params []interface{}) interface{}
	Params []interface{}
	Ctx    context.Context
}

// NewDefaultRunnable 实例化默认执行体
func NewDefaultRunnable(ctx context.Context,
	taskMethod func(ctx context.Context, params []interface{}) interface{},
	taskName string,
	taskParam []interface{}) *DefaultRunnable {
	object := &DefaultRunnable{
		Method: taskMethod,
		Params: taskParam,
		Ctx:    ctx,
	}
	object.ID = atomic.AddInt64(&nextRunnableID, 1)
	object.Name = taskName
	if 0 >= len(object.Name) {
		object.Name = "Unknown"
	}
	return object
}

// GetID 获取ID
func (object *DefaultRunnable) GetID() string {
	return fmt.Sprintf(`%s-%d`, object.Name, object.ID)
}

// GetParams 获取参数
func (object *DefaultRunnable) GetParams() []interface{} {
	return object.Params
}

// GetCtx 获取上下文
func (object *DefaultRunnable) GetCtx() context.Context {
	return object.Ctx
}

// Start 执行
func (object *DefaultRunnable) Run(params []interface{}) interface{} {
	return object.Method(object.Ctx, params)
}

// 字符串描述
func (object *DefaultRunnable) String() string {
	return object.GetID()
}

// 崩溃处理器
type PanicHandler func(r interface{})

// RoutinePool 协程池
type RoutinePool struct {
	panicHandler  PanicHandler       // 崩溃处理器
	stopFlag      int32              // 停止标志
	minRoutine    int                // 最小协程持有
	minPoolSize   int32              // 协程池最小大小
	maxPoolSize   int32              // 协程池最大大小
	ctx           context.Context    // 上下文
	cancel        context.CancelFunc // 撤销方法
	wg            sync.WaitGroup     // 等待组
	taskQueue     chan Runnable      // 任务队列
	taskQueueSize int32              // 任务队列大小
}

// New 工厂方法
func New(minPoolSize, maxPoolSize int) *RoutinePool {
	object := &RoutinePool{
		stopFlag:    0,
		minRoutine:  minPoolSize,
		minPoolSize: int32(minPoolSize),
		maxPoolSize: int32(maxPoolSize),
		taskQueue:   make(chan Runnable, defaultTaskQueueSize),
	}
	if 0 == object.maxPoolSize {
		object.maxPoolSize = math.MaxInt32
	}
	object.ctx, object.cancel = context.WithCancel(context.Background())
	for i := 0; i < minPoolSize; i++ {
		go object.loop()
	}
	return object
}

// SetPanicHandler 设置崩溃处理器
func (object *RoutinePool) SetPanicHandler(handler PanicHandler) *RoutinePool {
	object.panicHandler = handler
	return object
}

// 执行任务
func (object *RoutinePool) doWork(task Runnable) {
	defer func() {
		if r := recover(); nil != r {
			// 出Bug了，优雅关闭吧
			if nil != object.panicHandler {
				object.panicHandler(r)
			} else {
				glog.Error(r)
				debug.PrintStack()
				app.GetInstance().Shutdown()
			}
		}
	}()
	if app_cfg.GetInstance().Debug {
		glog.Infof("task %s begin", task)
	}
	task.Run(task.GetParams())
	if app_cfg.GetInstance().Debug {
		glog.Infof("task %s end", task)
	}
}

// 循环
func (object *RoutinePool) loop() {
	object.wg.Add(1)
loop:
	for {
		select {
		case <-object.ctx.Done():
			break loop
		case task := <-object.taskQueue:
			atomic.AddInt32(&object.minPoolSize, -1)
			object.doWork(task)
			if int32(object.minRoutine) < atomic.LoadInt32(&object.minPoolSize) {
				break loop
			}
			atomic.AddInt32(&object.minPoolSize, 1)
			atomic.AddInt32(&object.taskQueueSize, -1)
		}
	}
	object.wg.Done()
}

// CommitRunnable 提交任务
func (object *RoutinePool) CommitRunnable(runnable Runnable) (err error) {
	if object.IsStopped() {
		err = ErrRoutinePoolClosed
		return
	}
	taskQueueSize := atomic.AddInt32(&object.taskQueueSize, 1)
	if taskQueueSize >= atomic.LoadInt32(&object.minPoolSize) &&
		atomic.LoadInt32(&object.maxPoolSize) > taskQueueSize /*限定池的大小*/ {
		go object.loop()
	}
	glog.Info("post task ", runnable)
	object.taskQueue <- runnable
	return nil
}

// CommitTask 提交任务
func (object *RoutinePool) CommitTask(task func(ctx context.Context, params []interface{}) interface{},
	taskName string,
	params ...interface{}) (err error) {
	return object.CommitRunnable(NewDefaultRunnable(object.ctx, task, taskName, params))
}

// 是否已经停止
func (object *RoutinePool) IsStopped() bool {
	return 1 == atomic.LoadInt32(&object.stopFlag)
}

// stop 停止
func (object *RoutinePool) Stop() {
	object.cancel()
	object.wg.Wait()
	// 由调用者清空任务队列
	atomic.StoreInt32(&object.stopFlag, 1)
	for 0 < len(object.taskQueue) {
		object.doWork(<-object.taskQueue)
		if int32(object.minRoutine) < atomic.LoadInt32(&object.minPoolSize) {
			atomic.AddInt32(&object.minPoolSize, -1)
		}
	}
}

// GetInstance 获取单例
func GetInstance() *RoutinePool {
	routinePoolOnce.Do(func() {
		routinePoolInstance = New(16, 0)
	})
	return routinePoolInstance
}
