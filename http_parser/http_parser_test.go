package http_parser

import "testing"

func TestDefaultHttpParser(t *testing.T) {
	httpParser := NewDefaultHTTPParser()
	status := httpParser.SetMethodHook(func(value []byte) {
		t.Log("method:", string(value))
	}).SetURLHook(func(value []byte) {
		t.Log("url:", string(value))
	}).SetProtocolHook(func(value []byte) {
		t.Log("protocol:", string(value))
	}).SetHeaderHook(func(key, value []byte) {
		t.Log("header:", string(key), string(value))
	}).SetBodyHook(func(value []byte) {
		t.Log("body:", string(value))
	}).SetChunkedHook(func(value []byte) {
	}).Process([]byte("GET / HTTP/1.1\r\nHost: localhost\r\nContent-Length: 5\r\nConnection: close\r\n\r\nhello"))
	if HTTPParserStatusOK != status {
		t.Error("")
		return
	}
}
