package service

import (
	"context"
	"errors"
	"github.com/buaazp/fasthttprouter"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/priority_define"
	"github.com/valyala/fasthttp"
	"net/http"
	"reflect"
)

// 错误定义
var (
	ErrServerAddress = errors.New("server address error") // 服务器地址错误
)

// HTTPServiceConfig HTTP服务器配置
type HTTPServiceConfig struct {
	HTTPAddr           string // HTTP地址
	HTTPStaticLocation string // HTTP静态文件路径(URL)
	HTTPStaticPath     string // HTTP静态文件本地路径
	HTTPSAddr          string // HTTPS地址
	HTTPSCertPath      string // HTTP证书
	HTTPSKeyPath       string // HTTP Key
}

// HTTPService HTTP服务
type HTTPService struct {
	*HTTPServiceConfig
	srv    *fasthttp.Server
	Router *fasthttprouter.Router
}

// NewHTTPService 工厂方法
func NewHTTPService(config *HTTPServiceConfig) *HTTPService {
	object := &HTTPService{Router: fasthttprouter.New()}
	object.HTTPServiceConfig = config
	object.srv = &fasthttp.Server{
		Handler: object.Router.Handler,
	}
	event_bus.GetInstance().MountingOnce(reflect.TypeOf(&event.AppShutdownEvent{}),
		"HTTPServiceConfig",
		func(ctx context.Context, param interface{}) {
			if priority_define.HTTPServiceShutdownPriority !=
				param.(*event.AppShutdownEvent).ShutdownPriority {
				return
			}
			object.Shutdown()
			glog.Info("HTTPService done")
		})
	return object
}

// Shutdown 关闭
func (object *HTTPService) Shutdown() (err error) {
	err = object.srv.Shutdown()
	return
}

// Start 启动
func (object *HTTPService) Start() (err error) {
	if 0 >= len(object.HTTPAddr) && 0 >= len(object.HTTPSAddr) {
		err = ErrServerAddress
		return
	}

	if 0 < len(object.HTTPStaticLocation) && 0 < len(object.HTTPStaticPath) {
		var staticLocation string
		if '/' == object.HTTPStaticLocation[len(object.HTTPStaticLocation)-1] {
			staticLocation = object.HTTPStaticLocation + "*filepath"
		} else {
			staticLocation = object.HTTPStaticLocation + "/*filepath"
		}
		object.Router.ServeFiles(staticLocation, object.HTTPStaticPath)
	}
	if 0 < len(object.HTTPSAddr) {
		// HTTPS
		err = object.srv.ListenAndServeTLS(object.HTTPSAddr, object.HTTPSCertPath, object.HTTPSKeyPath)
	}
	// HTTP
	if 0 < len(object.HTTPAddr) {
		err = object.srv.ListenAndServe(object.HTTPAddr)
		if http.ErrServerClosed == err {
			err = nil
		}
	}
	return
}
