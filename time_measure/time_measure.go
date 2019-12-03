package time_measure

import (
	"fmt"
	"time"
)

// TimeMeasure 时间测量工具
type TimeMeasure struct {
	start    int64
	delta    int64
	funcName string
}

// NewTimeMeasure 工厂方法
func NewTimeMeasure(funcName string) *TimeMeasure {
	return &TimeMeasure{
		start:    time.Now().UnixNano(),
		funcName: funcName,
	}
}

// SetFuncName 设置函数名
func (object *TimeMeasure) SetFuncName(funcName string) *TimeMeasure {
	object.funcName = funcName
	return object
}

// Stop 停止
func (object *TimeMeasure) Stop() int64 {
	object.delta = time.Now().UnixNano() - object.start
	return object.delta
}

// Reset 重置
func (object *TimeMeasure) Reset() {
	object.start = time.Now().UnixNano()
}

// ToString 字符串描述
func (object *TimeMeasure) String() string {
	return fmt.Sprintf("func: %s, use: %d nanoseconds", object.funcName, object.delta)
}
