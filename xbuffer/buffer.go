package xbuffer

import (
	"bytes"
	"encoding/binary"
)

// Buffer 对象
type Buffer bytes.Buffer

// WriteFixData 写固定长度数据
func (object *Buffer) WriteFixData(v interface{}, bigEndian bool) (err error) {
	byteOrder := binary.ByteOrder(binary.BigEndian)
	if !bigEndian {
		byteOrder = binary.LittleEndian
	}
	err = binary.Write((*bytes.Buffer)(object), byteOrder, v)
	return
}
