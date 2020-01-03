package byte_buf

import (
	"bytes"
	"math"
	"testing"
)

func TestByteBufDefaultCap(t *testing.T) {
	buf := New()
	if 1<<12 != buf.initCap || 1<<12 != len(buf.buf) {
		t.Error("init cap error")
		return
	}

	if math.MaxInt32 != buf.maxCap {
		t.Error("max cap error")
		return
	}

	if 0 != buf.ReaderIndex() ||
		0 != buf.WriterIndex() ||
		0 != buf.mrIndex ||
		0 != buf.mwIndex {
		t.Error("index error")
		return
	}
}

func TestByteBufCap(t *testing.T) {
	buf := New(InitCapOption(16), MaxCapOption(32))
	if 16 != buf.initCap || 32 != buf.maxCap {
		t.Error("init cap error")
		return
	}
}

func TestWrappedOption(t *testing.T) {
	arr := make([]byte, 32)
	buf := New(WrappedOption(arr))
	if 32 != buf.initCap || math.MaxInt32 != buf.maxCap {
		t.Error("init or max cap error")
		return
	}
}

func TestByteBuf_DiscardReadBytes(t *testing.T) {
	buf := New()
	buf.WriteBytes([]byte("HELLO"))
	if 5 != buf.ReadableBytes() {
		t.Error("ReadableBytes")
		return
	}

	if !bytes.Equal(buf.GetBytes(1024), []byte("HELLO")) {
		t.Error("GetBytes")
		return
	}

	if 0 != buf.DiscardReadBytes().ReadableBytes() {
		t.Error("DiscardReadBytes")
		return
	}
}

func TestWriteRead(t *testing.T) {
	buf := New()
	buf.WriteBool(true)
	if 1 != buf.ReadableBytes() ||
		!buf.GetBool() ||
		0 != buf.ReadableBytes() {
		t.Error("ReadableBytes || GetBool")
		return
	}

	buf.WriteBool(false)
	if 1 != buf.ReadableBytes() ||
		buf.GetBool() ||
		0 != buf.ReadableBytes() {
		t.Error("ReadableBytes || GetBool")
		return
	}

	buf.WriteByte(byte('\r'))
	if 1 != buf.ReadableBytes() ||
		'\r' != buf.GetByte() ||
		0 != buf.ReadableBytes() {
		t.Error("ReadableBytes || GetByte")
		return
	}

	buf.WriteUint8(uint8('\n'))
	if 1 != buf.ReadableBytes() ||
		'\n' != buf.GetUint8() ||
		0 != buf.ReadableBytes() {
		t.Error("ReadableBytes || GetUint8")
		return
	}

	buf.WriteUint16(uint16(0xefff))
	if 2 != buf.ReadableBytes() ||
		0xefff != buf.GetUint16() ||
		0 != buf.ReadableBytes() {
		t.Error("ReadableBytes || GetUint16")
		return
	}

	buf.WriteUint16LE(uint16(0xefff))
	if 2 != buf.ReadableBytes() ||
		0xefff != buf.GetUint16LE() ||
		0 != buf.ReadableBytes() {
		t.Error("ReadableBytes || GetUint16")
		return
	}

	buf.WriteMedium(0xfffffe)
	if 3 != buf.ReadableBytes() ||
		0xfffffe != buf.GetMedium() ||
		0 != buf.ReadableBytes() {
		t.Error("ReadableBytes || GetMedium")
		return
	}

	buf.WriteMediumLE(0xfffffe)
	if 3 != buf.ReadableBytes() ||
		0xfffffe != buf.GetMediumLE() ||
		0 != buf.ReadableBytes() {
		t.Error("ReadableBytes || GetMediumLE")
		return
	}

	buf.WriteUint32(0xfefefefe)
	if 4 != buf.ReadableBytes() ||
		0xfefefefe != buf.GetUint32() ||
		0 != buf.ReadableBytes() {
		t.Error("ReadableBytes || GetUint32")
		return
	}

	buf.WriteUint32LE(0xfefefefe)
	if 4 != buf.ReadableBytes() ||
		0xfefefefe != buf.GetUint32LE() ||
		0 != buf.ReadableBytes() {
		t.Error("ReadableBytes || GetUint32LE")
		return
	}

	buf.WriteUint64(0xfefefefefefefefe)
	if 8 != buf.ReadableBytes() ||
		0xfefefefefefefefe != buf.GetUint64() ||
		0 != buf.ReadableBytes() {
		t.Error("ReadableBytes || GetUint64")
		return
	}

	buf.WriteUint64LE(0xfefefefefefefefe)
	if 8 != buf.ReadableBytes() ||
		0xfefefefefefefefe != buf.GetUint64LE() ||
		0 != buf.ReadableBytes() {
		t.Error("ReadableBytes || GetUint64")
		return
	}

	buf.WriteFloat32(0.1)
	if 4 != buf.ReadableBytes() ||
		0.1 != buf.GetFloat32() ||
		0 != buf.ReadableBytes() {
		t.Error("ReadableBytes || GetFloat32")
		return
	}

	buf.WriteFloat32LE(0.1)
	if 4 != buf.ReadableBytes() ||
		0.1 != buf.GetFloat32LE() ||
		0 != buf.ReadableBytes() {
		t.Error("ReadableBytes || GetFloat32")
		return
	}

	buf.WriteFloat64(0.1)
	if 8 != buf.ReadableBytes() ||
		0.1 != buf.GetFloat64() ||
		0 != buf.ReadableBytes() {
		t.Error("ReadableBytes || GetFloat64")
		return
	}

	buf.WriteFloat64LE(0.1)
	if 8 != buf.ReadableBytes() ||
		0.1 != buf.GetFloat64LE() ||
		0 != buf.ReadableBytes() {
		t.Error("ReadableBytes || GetFloat64LE")
		return
	}
}

func TestByteBuf_Skip(t *testing.T) {
	buf := New()
	buf.WriteBytes([]byte("\r\n"))
	buf.WriteBytes([]byte("GET / HTTP/1.1/r/n"))
	buf.Skip('\r').Skip('\n')
	if 'G' != buf.GetByte() {
		t.Error("skip")
	}
}

func TestByteBuf_TakeUntil(t *testing.T) {
	buf := New()
	buf.WriteBytes([]byte("GET / HTTP/1.1\r\n"))
	if !bytes.Equal([]byte("GET"), buf.TakeUntil(' ', true)) {
		t.Error("TakeUntil")
		return
	}

	buf.Skip(' ')
	if !bytes.Equal([]byte("/"), buf.TakeUntil(' ', true)) {
		t.Error("TakeUntil")
		return
	}

	buf.Skip(' ')
	if !bytes.Equal([]byte("HTTP/1.1"), buf.TakeUntil('\r', true)) {
		t.Error("TakeUntil")
		return
	}

	if 0 != buf.Skip('\r').Skip('\n').ReadableBytes() {
		t.Error("Skip")
		return
	}

	buf.DiscardReadBytes()
	if 0 != buf.ReaderIndex() || 0 != buf.WriterIndex() {
		t.Error("DiscardReadBytes")
		return
	}
}
