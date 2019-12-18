package asset

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/intelligentfish/gogo/app"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/routine_pool"
	"github.com/jinzhu/gorm"
	"testing"
	"time"
)

func TestUserAsset(t *testing.T) {
	flag.Parse()

	db, err := gorm.Open("mysql", "root:dev@(172.25.73.216:3306)/dev1?charset=utf8&parseTime=True&loc=Local")
	if nil != err {
		t.Error(err)
		return
	}
	defer db.Close()

	db.AutoMigrate(&UserAsset{})

	c := redis.NewClient(&redis.Options{
		Addr:     "172.25.73.216:6379",
		Password: "",
		DB:       0,
	})
	_, err = c.Ping().Result()
	if nil != err {
		t.Error(err)
		return
	}

	asset := &UserAsset{
		Nickname: "neo",
		Avatar:   "http://www.google.com",
		Balance:  10000000,
	}
	asset.ID = 1
	if err = Util.UpdateAsset(c, asset); nil != err {
		t.Error(err)
		return
	}

	routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case <-time.After(10 * time.Microsecond):
				asset := &UserAsset{}
				asset.ID = 1
				err := Util.GetAsset(c, asset)
				if nil != err {
					t.Error(err)
					break loop
				}
				asset.Balance--
				err = Util.UpdateAsset(c, asset)
				if nil != err {
					t.Error(err)
					break loop
				}
			}
		}
		return nil
	}, "SubUserBalanceJob")
	routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case <-time.After(1 * time.Second):
				userAssets, err := Util.GetDirtyUserAsset(c, &UserAsset{})
				if nil != err {
					t.Error(err)
					break loop
				}
				for _, asset := range userAssets {
					fmt.Println(asset)
					var count int
					err = db.Model(&UserAsset{}).Where("id = ?", asset.(*UserAsset).ID).Count(&count).Error
					if nil != err {
						t.Error(err)
						break loop
					}
					if 0 >= count {
						err = db.Create(asset).Error
						if nil != err {
							t.Error(err)
							break loop
						}
						continue
					}
					err = db.Save(asset).Error
					if nil != err {
						t.Error(err)
						break loop
					}
				}
			}
		}
		return nil
	}, "UserAssetSyncJob")

	app.GetInstance().AddShutdownHook(
		event_bus.GetInstance().NotifyAllComponentShutdown, // 通知所有组件关闭
		event_bus.GetInstance().Stop,                       // 停止事件总线
		routine_pool.GetInstance().Stop,                    // 停止协程池
	).WaitShutdown()
}
