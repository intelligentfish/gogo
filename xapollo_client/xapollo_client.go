package xapollo_client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/priority_define"
	"github.com/intelligentfish/gogo/routine_pool"
	"github.com/valyala/fasthttp"
	"net/http"
	"reflect"
	"sync"
	"time"
)

const (
	configPullInterval = time.Second // 配置拉取间隔
)

var (
	apolloClientOnce     sync.Once     // 一次执行
	apolloClientInstance *ApolloClient // 单实例
)

// ConfigChangeNotification 配置改变通知
type ConfigChangeNotification struct {
	AppID          string            `json:"appId"`          // 程序ID
	Cluster        string            `json:"cluster"`        // 集群
	NamespaceName  string            `json:"namespaceName"`  // 命名空间
	Configurations map[string]string `json:"configurations"` // 配置
	ReleaseKey     string            `json:"releaseKey"`     // 发行版本key
}

// ApolloClient Apollo客户端
type ApolloClient struct {
	ConfigServerAddr string // 配置服务器地址
	AppID            string // 应用id
	Cluster          string // 集群
	Namespace        string // 命名空间
	ReleaseKey       string // 发行Key
}

// 工厂方法
func newApolloClient() *ApolloClient {
	return &ApolloClient{}
}

// PullConfig 拉取配置
func (object *ApolloClient) PullConfig() (ccn *ConfigChangeNotification, err error) {
	url := fmt.Sprintf("http://%s/configs/%s/%s/%s?releaseKey=%s",
		object.ConfigServerAddr,
		object.AppID,
		object.Cluster,
		object.Namespace,
		object.ReleaseKey)
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	req.SetRequestURI(url)
	req.Header.SetMethod(http.MethodGet)
	err = fasthttp.Do(req, res)
	if nil != err || http.StatusOK != res.StatusCode() {
		return
	}

	err = json.Unmarshal(res.Body(), &ccn)
	if nil == err {
		object.ReleaseKey = ccn.ReleaseKey
	}
	return
}

// InitializeWithTimeout 初始化
func (object *ApolloClient) Initialize(configServerAddr,
	appID,
	cluster,
	namespace string) *ApolloClient {
	object.ConfigServerAddr = configServerAddr
	object.AppID = appID
	object.Cluster = cluster
	object.Namespace = namespace
	return object
}

// RegisterConfigChangeNotification 注册配置改变事件
func (object *ApolloClient) RegisterConfigChangeNotification(callback func(param interface{})) {
	event_bus.GetInstance().Mounting(reflect.TypeOf(&ConfigChangeNotification{}), callback)
}

// Start 启动
func (object *ApolloClient) Start() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	event_bus.GetInstance().MountingOnce(reflect.TypeOf(&event.AppShutdownEvent{}),
		"ApolloClient",
		func(param interface{}) {
			if priority_define.ConfigClientShutdownPriority !=
				param.(*event.AppShutdownEvent).ShutdownPriority {
				return
			}

			cancel()
			wg.Wait()

			glog.Info("ApolloClient done")
		})
	routine_pool.GetInstance().PostTask(func(params []interface{}) interface{} {
		defer wg.Done()

	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case <-time.After(configPullInterval):
				ccn, err := object.PullConfig()
				if nil != err {
					glog.Error(err)
					continue
				}

				if nil == ccn {
					continue
				}

				event_bus.GetInstance().SyncNotify(reflect.TypeOf(&ConfigChangeNotification{}), ccn)
			}
		}
		return nil
	}, "ApolloClient")
}

// GetApolloClientInstance 获取实例
func GetApolloClientInstance() *ApolloClient {
	apolloClientOnce.Do(func() {
		apolloClientInstance = newApolloClient()
	})
	return apolloClientInstance
}
