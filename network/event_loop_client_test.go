package network

import (
	"flag"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/byte_buf"
	"net"
	"sync"
	"testing"
)

func TestEventLoopClient(t *testing.T) {
	flag.Parse()
	flag.Set("v", "0")
	flag.Set("logtostderr", "true")

	wg := sync.WaitGroup{}
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		conn, err := net.Dial("tcp", "127.0.0.1:10080")
		if nil != err {
			glog.Error(err)
			wg.Done()
			return
		}

		if _, err = conn.Write([]byte("hello")); nil != err {
			glog.Error(err)
			wg.Done()
			return
		}

		buf := byte_buf.GetPoolInstance().Borrow(byte_buf.InitCapOption(1 << 13))
		if _, err = conn.Read(buf.Internal()[buf.WriterIndex():buf.InitCap()]); nil != err {
			glog.Error(err)
			wg.Done()
			return
		}

		glog.Info(string(buf.Internal()[buf.ReaderIndex():buf.WriterIndex()]))
		byte_buf.GetPoolInstance().Return(buf.DiscardAllBytes())
		wg.Done()
	}
	wg.Wait()
	glog.Info("event loop client done")
}
