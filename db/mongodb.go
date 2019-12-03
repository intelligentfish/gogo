package db

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/app_cfg"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/priority_define"
	"gopkg.in/mgo.v2"
	"reflect"
	"sync"
	"time"
)

// 常量
const (
	defaultPoolSizeLimit = 1024 // 连接池大小
)

var (
	mongoDBOnce     sync.Once // 执行一次
	mongoDBInstance *MongoDB  // 单实例
)

// MongoDB 数据库
type MongoDB struct {
	session *mgo.Session    // 会话
	wg      *sync.WaitGroup // 等待组
}

// WithCollection 获取集合
func (object *MongoDB) WithCollection(db, c string, callback func(c *mgo.Collection)) {
	s := object.session.Clone()
	defer func() {
		if r := recover(); nil != r {
			switch rt := r.(type) {
			case error:
				glog.Error(rt)
			default:
				glog.Error(fmt.Sprint(rt))
			}
		}
		s.Close()
		object.wg.Done()
	}()

	object.wg.Add(1)
	callback(s.DB(db).C(c))
}

// WithDefaultDBCollection 获取默认数据库集合
func (object *MongoDB) WithDefaultDBCollection(c string, callback func(c *mgo.Collection)) {
	object.WithCollection(app_cfg.GetInstance().MongoDBCfg.DbName, c, callback)
}

// InitializeWithTimeout 初始化
func (object *MongoDB) InitializeWithTimeout(timeout time.Duration) (err error) {
	object.session, err = mgo.DialWithTimeout(app_cfg.GetInstance().MongoDBCfg.MongoDBUrl, timeout)
	if nil == err {
		object.session.SetMode(mgo.Monotonic, true)
		object.session.SetPoolLimit(defaultPoolSizeLimit)
		event_bus.GetInstance().MountingOnce(reflect.TypeOf(&event.AppShutdownEvent{}),
			"MongoDB",
			func(param interface{}) {
				if priority_define.DBShutdownPriority != param.(*event.AppShutdownEvent).ShutdownPriority {
					return
				}

				object.wg.Wait()
				object.session.Close()

				glog.Info("MongoDB done")
			}).SyncNotify(reflect.TypeOf(&event.MongoDBInitializeEvent{}),
			&event.MongoDBInitializeEvent{}) // 通知数据库初始化完成
	}
	return
}

// GetMongoDBInstance 获取单例
func GetMongoDBInstance() *MongoDB {
	mongoDBOnce.Do(func() {
		mongoDBInstance = &MongoDB{wg: &sync.WaitGroup{}}
	})
	return mongoDBInstance
}
