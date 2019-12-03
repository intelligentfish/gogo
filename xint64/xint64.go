package xint64

import "strconv"

// Int64 扩展Int64
type Int64 int64

// ToString 转换为字符串
func (object Int64) ToString() string {
	return strconv.FormatInt(int64(object), 10)
}
