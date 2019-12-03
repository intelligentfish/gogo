package comparable

// Comparable 可比较对象
type Comparable interface {
	CompareTo(other interface{}) int
}
