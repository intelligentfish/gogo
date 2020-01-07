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
	GetID() string            // ID
	GetParams() []interface{} // 获取参数
	GetCtx() context.Context  // 获取上下文
	Run(params []interface{}) // 执行
}

// DefaultRunnable 默认实现
type DefaultRunnable struct {
	ID     int64
	Name   string
	Method func(ctx context.Context, params []interface{})
	Params []interface{}
	Ctx    context.Context
}

// NewDefaultRunnable 实例化默认执行体
func NewDefaultRunnable(ctx context.Context,
	taskMethod func(ctx context.Context, params []interface{}),
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
func (object *DefaultRunnable) Run(params []interface{}) {
	object.Method(object.Ctx, params)
}

// 字符串描述
func (object *DefaultRunnable) String() string {
	return object.GetID()
}

// 崩溃处理器
type PanicHandler func(r interface{})

// RoutinePool 协程池
type RoutinePool struct {
	panicHandler    PanicHandler       // 崩溃处理器
	stopFlag        int32              // 停止标志
	currentPoolSize int32              // 协程池当前大小
	minPoolSize     int32              // 协程池最小大小
	maxPoolSize     int32              // 协程池最大大小
	ctx             context.Context    // 上下文
	cancel          context.CancelFunc // 撤销方法
	wg              sync.WaitGroup     // 等待组
	taskQueue       chan Runnable      // 任务队列
	taskQueueSize   int32              // 任务队列大小
}

// Option 携程池选项
type Option func(object *RoutinePool)

// MinPoolSizeOption 最小池大小选项
func MinPoolSizeOption(poolSize int) Option {
	return func(object *RoutinePool) {
		atomic.StoreInt32(&object.minPoolSize, int32(poolSize))
		atomic.StoreInt32(&object.currentPoolSize, int32(poolSize))
	}
}

// MaxPoolSizeOption 最大池大小选项
func MaxPoolSizeOption(poolSize int) Option {
	return func(object *RoutinePool) {
		atomic.StoreInt32(&object.maxPoolSize, int32(poolSize))
	}
}

// New 工厂方法
func New(options ...Option) *RoutinePool {
	defaultPoolSize := int32(16)
	object := &RoutinePool{
		stopFlag:        0,
		minPoolSize:     defaultPoolSize,
		currentPoolSize: defaultPoolSize,
		maxPoolSize:     math.MaxInt16,
		taskQueue:       make(chan Runnable, defaultTaskQueueSize),
	}
	for _, option := range options {
		option(object)
	}
	if 0 == object.maxPoolSize {
		object.maxPoolSize = math.MaxInt32
	}
	object.ctx, object.cancel = context.WithCancel(context.Background())
	for i := 0; i < int(object.minPoolSize); i++ {
		go object.loop()
	}
	return object
}

// SetOption 设置选项
func (object *RoutinePool) SetOption(options ...Option) *RoutinePool {
	for _, option := range options {
		option(object)
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
	atomic.AddInt32(&object.currentPoolSize, 1)
	object.wg.Add(1)
loop:
	for {
		select {
		case <-object.ctx.Done():
			break loop
		case task := <-object.taskQueue:
			object.doWork(task)
			atomic.AddInt32(&object.taskQueueSize, -1)
			if atomic.LoadInt32(&object.currentPoolSize) > atomic.LoadInt32(&object.minPoolSize) {
				break loop
			}
		}
	}
	object.wg.Done()
	atomic.AddInt32(&object.currentPoolSize, -1)
}

// CommitRunnable 提交任务
func (object *RoutinePool) CommitRunnable(runnable Runnable) (err error) {
	if object.IsStopped() {
		err = ErrRoutinePoolClosed
		return
	}
	if atomic.AddInt32(&object.taskQueueSize, 1) >= atomic.LoadInt32(&object.minPoolSize) &&
		atomic.LoadInt32(&object.maxPoolSize) > atomic.LoadInt32(&object.taskQueueSize) /*限定池的大小*/ {
		go object.loop()
	}
	object.taskQueue <- runnable
	return nil
}

// CommitTask 提交任务
func (object *RoutinePool) CommitTask(task func(ctx context.Context, params []interface{}),
	taskName string,
	params ...interface{}) (err error) {
	return object.CommitRunnable(NewDefaultRunnable(object.ctx, task, taskName, params))
}

// 是否已经停止
func (object *RoutinePool) IsStopped() bool {
	return 1 == atomic.LoadInt32(&object.stopFlag)
}

// Stop 停止
func (object *RoutinePool) Stop() {
	object.cancel()
	object.wg.Wait()
	// 由调用者清空任务队列
	atomic.StoreInt32(&object.stopFlag, 1)
	for 0 < len(object.taskQueue) {
		object.doWork(<-object.taskQueue)
		if object.currentPoolSize < atomic.LoadInt32(&object.minPoolSize) {
			atomic.AddInt32(&object.currentPoolSize, -1)
		}
	}
}

// TaskQueueSize 任务队列大小
func (object *RoutinePool) TaskQueueSize() int32 {
	return atomic.LoadInt32(&object.taskQueueSize)
}

// PoolSize 池大小
func (object *RoutinePool) PoolSize() int32 {
	return atomic.LoadInt32(&object.currentPoolSize)
}

// GetInstance 获取单例
func GetInstance() *RoutinePool {
	routinePoolOnce.Do(func() {
		routinePoolInstance = New(MinPoolSizeOption(1024),
			MaxPoolSizeOption(math.MaxInt16))
	})
	return routinePoolInstance
}
