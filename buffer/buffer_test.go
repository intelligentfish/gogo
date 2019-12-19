package buffer

import "testing"

func TestBuffer(t *testing.T) {
	buf := NewBuffer(0)
	for i := 0; i < 10; i++ {
		buf.WriteUint32(uint32(i))
	}
	if 0 != buf.GetReadIndex() || 40 != buf.GetWriteIndex() || 40 != buf.ReadableBytes() {
		t.Error("")
		return
	}

	for i := 0; i < 10; i++ {
		if uint32(i) != buf.ReadUint32() {
			t.Error("")
			return
		}
	}

	if 40 != buf.GetReadIndex() || buf.GetReadIndex() != buf.GetWriteIndex() || 0 != buf.ReadableBytes() || !buf.IsEmpty() {
		t.Error("")
		return
	}

	buf.DiscardReadBytes()
	if 0 != buf.GetReadIndex() || 0 != buf.GetWriteIndex() || cap(buf.Internal) != buf.WriteableBytes() || !buf.IsEmpty() {
		t.Error("")
		return
	}
}
