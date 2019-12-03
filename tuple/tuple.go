package tuple

import "fmt"

// Tuple 二元组
type Tuple struct {
	A int
	B int
}

// String 字符串描述
func (object *Tuple) String() string {
	return fmt.Sprintf("(%d,%d)", object.A, object.B)
}
