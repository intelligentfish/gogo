package health_checker

import (
	"context"
	"flag"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/app"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/routine_pool"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestHealthChecker(t *testing.T) {
	flag.Parse()
	glog.Infof("app id: %d", app.GetInstance().PID)
	event_bus.GetInstance().Mounting(reflect.TypeOf(&event.HealthCheckRequest{}),
		func(ctx context.Context, param interface{}) {
			req := param.(*event.HealthCheckRequest)
			if 0 < len(req.UniqueName) && req.UniqueName != "Test" {
				return
			}
			sec := rand.Int31n(20)
			if 0 == sec {
				sec = 1
			}
			glog.Errorf("Test sleep %d seconds", sec)
			select {
			case <-ctx.Done():
			case <-time.After(time.Duration(sec) * time.Second):
			}
			event_bus.GetInstance().Notify(reflect.TypeOf(&event.HealthCheckReply{}),
				&event.HealthCheckReply{
					UniqueName: "Test",
					OK:         true,
				})
		})
	GetHealthCheckerInstance().Start()
	app.GetInstance().AddShutdownHook(
		event_bus.GetInstance().NotifyAllComponentShutdown, // 通知所有组件关闭
		event_bus.GetInstance().Stop,                       // 停止事件总线
		routine_pool.GetInstance().Stop,                    // 停止协程池
	).WaitShutdown()
}
