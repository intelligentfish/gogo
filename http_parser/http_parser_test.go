package http_parser

import (
	"bytes"
	"github.com/intelligentfish/gogo/byte_buf"
	"testing"
)

func TestNew(t *testing.T) {
	byteBuf := byte_buf.New()
	byteBuf.WriteBytes([]byte("GET / HTTP/1.1\r\nConnect: close\r\nContent-Length: 5\r\n\r\nHello"))
	parser := New(ByteBufOption(byteBuf))
	if ParseResultOK != parser.Parse() {
		t.Error("Parse")
		return
	}

	if "GET" != parser.GetMethod() {
		t.Error("GetMethod")
		return
	}

	if "/" != parser.GetURI() {
		t.Error("GetURI")
		return
	}

	if "HTTP/1.1" != parser.GetVersion() {
		t.Error("GetVersion")
		return
	}

	if "close" != parser.GetHeader("Connect") {
		t.Error()
	}

	if 5 != parser.GetContentLength() ||
		"" != parser.GetContentType() {
		t.Error("GetContentLength or GetContentType")
		return
	}

	start, end := parser.GetBodyRange()
	if 0 != start && 5 != end {
		t.Error("GetBodyRange")
		return
	}

	if bytes.Equal([]byte("hello"), byteBuf.Slice(start, end-start)) {
		t.Error("Body")
		return
	}

	if bytes.Equal([]byte("hello"), byteBuf.Slice(start, parser.GetContentLength())) {
		t.Error("Body")
		return
	}
}

func TestParseStream(t *testing.T) {
	byteBuf := byte_buf.New()
	raw := []byte("GET / HTTP/1.1\r\nConnect: close\r\nContent-Length: 5\r\n\r\nHello")
	parser := New(ByteBufOption(byteBuf))
	for {
		parseResult := parser.Parse()
		if ParseResultContinue != parseResult {
			if ParseResultOK != parseResult {
				t.Error("parse")
				return
			}
			break
		}
		byteBuf.WriteByte(raw[0])
		raw = raw[1:]
	}

	if "GET" != parser.GetMethod() {
		t.Error("GetMethod")
		return
	}

	if "/" != parser.GetURI() {
		t.Error("GetURI")
		return
	}

	if "HTTP/1.1" != parser.GetVersion() {
		t.Error("GetVersion")
		return
	}

	if "close" != parser.GetHeader("Connect") {
		t.Error()
	}

	if 5 != parser.GetContentLength() ||
		"" != parser.GetContentType() {
		t.Error("GetContentLength or GetContentType")
		return
	}

	start, end := parser.GetBodyRange()
	if 0 != start && 5 != end {
		t.Error("GetBodyRange")
		return
	}

	if bytes.Equal([]byte("hello"), byteBuf.Slice(start, end-start)) {
		t.Error("Body")
		return
	}

	if bytes.Equal([]byte("hello"), byteBuf.Slice(start, parser.GetContentLength())) {
		t.Error("Body")
		return
	}
}

func TestChunked(t *testing.T) {
	byteBuf := byte_buf.New()
	byteBuf.WriteBytes([]byte("GET / HTTP/1.1\r\nConnect: close\r\nTransfer-Encoding: chunked\r\n\r\n5\r\nHello0\r\n"))
	parser := New(ByteBufOption(byteBuf))
	if ParseResultOK != parser.Parse() {
		t.Error("Parse")
		return
	}

	if "GET" != parser.GetMethod() {
		t.Error("GetMethod")
		return
	}

	if "/" != parser.GetURI() {
		t.Error("GetURI")
		return
	}

	if "HTTP/1.1" != parser.GetVersion() {
		t.Error("GetVersion")
		return
	}

	if "close" != parser.GetHeader("Connect") {
		t.Error()
	}

	start, end := parser.GetBodyRange()
	if 0 != start && 5 != end {
		t.Error("GetBodyRange")
		return
	}

	if bytes.Equal([]byte("hello"), byteBuf.Slice(start, end-start)) {
		t.Error("Body")
		return
	}

	if bytes.Equal([]byte("hello"), byteBuf.Slice(start, parser.GetContentLength())) {
		t.Error("Body")
		return
	}
}

func TestChunkedStream(t *testing.T) {
	byteBuf := byte_buf.New()
	raw := []byte("GET / HTTP/1.1\r\nConnect: close\r\nTransfer-Encoding: chunked\r\n\r\n5\r\nHello0\r\n")
	parser := New(ByteBufOption(byteBuf))
	for {
		parseResult := parser.Parse()
		if ParseResultContinue != parseResult {
			if ParseResultOK != parseResult {
				t.Error("parse")
				return
			}
			break
		}
		byteBuf.WriteByte(raw[0])
		raw = raw[1:]
	}

	if "GET" != parser.GetMethod() {
		t.Error("GetMethod")
		return
	}

	if "/" != parser.GetURI() {
		t.Error("GetURI")
		return
	}

	if "HTTP/1.1" != parser.GetVersion() {
		t.Error("GetVersion")
		return
	}

	if "close" != parser.GetHeader("Connect") {
		t.Error()
	}

	start, end := parser.GetBodyRange()
	if 0 != start && 0 != end {
		t.Error("GetBodyRange")
		return
	}

	if bytes.Equal([]byte("hello"), byteBuf.Slice(start, end-start)) {
		t.Error("Body")
		return
	}

	if bytes.Equal([]byte("hello"), byteBuf.Slice(start, parser.GetContentLength())) {
		t.Error("Body")
		return
	}
}
