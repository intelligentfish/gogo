package triple

import "fmt"

// Triple 三元组
type Triple struct {
	A int
	B int
	C int
}

// String 字符串描述
func (object *Triple) String() string {
	return fmt.Sprintf("(%d,%d,%d)", object.A, object.B, object.C)
}
