package db

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/app_cfg"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/priority_define"
	"github.com/jinzhu/gorm"
	"reflect"
	"sync"
)

// 变量
var (
	mySQLOnce     sync.Once // 执行一次
	mySQLInstance *MySQL    // 单实例
)

// MySQL
type MySQL struct {
	DB *gorm.DB
}

// 工厂方法
func newMySQL() *MySQL {
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return app_cfg.GetInstance().MySQLDBCfg.Prefix + defaultTableName
	}
	return &MySQL{}
}

// 初始化
func (object *MySQL) Initialize() (err error) {
	cfg := app_cfg.GetInstance()
	object.DB, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.MySQLDBCfg.Username,
		cfg.MySQLDBCfg.Password,
		cfg.MySQLDBCfg.Host,
		cfg.MySQLDBCfg.DB,
	))
	if nil != err {
		return
	}
	event_bus.GetInstance().Mounting(reflect.TypeOf(&event.AppShutdownEvent{}),
		func(ctx context.Context, param interface{}) {
			if priority_define.DBShutdownPriority != param.(*event.AppShutdownEvent).ShutdownPriority {
				return
			}
			if err = object.DB.Close(); nil != err {
				glog.Error(err)
			} else {
				glog.Info("MySQL done")
			}
		}).SyncNotify(reflect.TypeOf(&event.MySQLInitializeEvent{}),
		&event.MySQLInitializeEvent{}) // 通知数据库初始化完成
	return
}

// GetMySQLInstance 获取单实例
func GetMySQLInstance() *MySQL {
	mySQLOnce.Do(func() {
		mySQLInstance = newMySQL()
	})
	return mySQLInstance
}
