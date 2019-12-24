package service

import (
	"context"
	"fmt"
	"github.com/intelligentfish/gogo/app"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/routine_pool"
	_ "net/http/pprof"
	"testing"
	"time"
)

func TestTCPService(t *testing.T) {
	NewTCPServiceWithCallback(func(session *TCPSession) (blocked bool) {
		session.SetName("server").AddCallback(
			func(session *TCPSession, chunk []byte) {
				//Echo
				session.Write(chunk)
			},
			func(session *TCPSession, isRead bool, err error) {
				if nil != err {
					t.Error(err)
				}
			})
		return false
	}).StartWithAddr(":10080")
	routine_pool.GetInstance().CommitTask(func(ctx context.Context, params []interface{}) {
		for i := 0; i < 10000; i++ {
			time.Sleep(1 * time.Millisecond)
			s := NewTCPSession().SetName("client").AddCallback(
				func(session *TCPSession, chunk []byte) {
					//fmt.Println(session.name, "-", session.ID, ",receive: ", string(chunk))
				}, func(session *TCPSession, isRead bool, err error) {
				})
			err := s.Connect("localhost:10080")
			if nil != err {
				t.Error(err)
				break
			}
			s.Start()
			routine_pool.GetInstance().CommitTask(func(ctx context.Context, params []interface{}) {
				for i := 0; i < 120; i++ {
					s.Write([]byte(fmt.Sprintf("the number is: %d", i)))
					time.Sleep(100 * time.Millisecond)
				}
				params[0].(*TCPSession).Stop()
				return
			}, "", s)
		}
		return
	}, "")
	routine_pool.GetInstance().CommitTask(func(ctx context.Context, params []interface{}) {
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case <-time.After(1 * time.Second):
				fmt.Println(fmt.Sprintf("(app id, task size, pool size) = (%d, %d, %d)",
					app.GetInstance().PID,
					routine_pool.GetInstance().TaskQueueSize(),
					routine_pool.GetInstance().PoolSize()))
			}
		}
		return
	}, "")
	app.GetInstance().
		EnablePPROF("").
		AddShutdownHook(event_bus.GetInstance().NotifyAllComponentShutdown,
			event_bus.GetInstance().Stop,
			routine_pool.GetInstance().Stop).
		WaitShutdown()
}
