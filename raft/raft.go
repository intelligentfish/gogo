package raft

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/auto_lock"
	"github.com/intelligentfish/gogo/event"
	"github.com/intelligentfish/gogo/event_bus"
	"github.com/intelligentfish/gogo/priority_define"
	"github.com/intelligentfish/gogo/routine_pool"
	"github.com/intelligentfish/gogo/xint64"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"time"
)

// Raft节点状态
type NodeState int32

const (
	NodeStateUnknown   = NodeState(iota) // 未知
	NodeStateFollower                    // 跟随者
	NodeStateCandidate                   // 候选者
	NodeStateLeader                      // 领导
)

// Node Raft节点
type Node struct {
	auto_lock.AutoLock
	myAddress          string   // 自己的地址
	leaderAddress      string   // 领导者的地址
	followerAddresses  []string // 追随者的地址
	otherNodeAddresses []string // 其他节点的地址
	term               int64    // 任期
	counter            int32    // 计数器
	overflowCount      int32    // 溢出计数值

	NodeState NodeState          // 节点状态
	ctx       context.Context    // 取消上下文
	cancel    context.CancelFunc // 取消方法
	wg        sync.WaitGroup     // 等待组
}

// NewNode() 工厂方法
func NewNode(myURL string, otherNodeURL ...string) *Node {
	object := &Node{
		myAddress:          myURL,
		otherNodeAddresses: otherNodeURL,
		NodeState:          NodeStateFollower,
	}
	object.overflowCount = 10
	object.ctx, object.cancel = context.WithCancel(context.Background())
	return object
}

// removeFollower 删除追随者
func (object *Node) removeFollower(address string) {
	for i := len(object.followerAddresses) - 1; i >= 0; i-- {
		if object.followerAddresses[i] == address {
			object.followerAddresses = append(object.followerAddresses[:i],
				object.followerAddresses[i+1:]...)
			break
		}
	}
}

// switchState  状态切换
func (object *Node) switchState() {
	object.wg.Add(1)
loop:
	for {
		select {
		case <-object.ctx.Done():
			// 停止
			break loop
		case <-time.After(1 * time.Second):
			glog.Infof("my address: %s, leader: %s\n", object.myAddress, object.leaderAddress)
			// 判断是否溢出
			flag := false
			object.WithLock(false,
				func() {
					object.counter++
					if object.overflowCount < object.counter {
						flag = true
						object.NodeState = NodeStateFollower
					}
				})
			if flag {
				continue
			}

			switch object.NodeState {
			case NodeStateFollower:
				object.WithLock(false,
					func() {
						if 0 >= object.counter {
							return
						}
						object.leaderAddress = ""
						object.followerAddresses = make([]string, 0)
						object.NodeState = NodeStateCandidate
					})
			case NodeStateCandidate:
				var addresses []string
				object.WithLock(false,
					func() {
						if 0 >= len(object.otherNodeAddresses) {
							return
						}
						object.counter = 0
						addresses = make([]string, len(object.otherNodeAddresses))
						copy(addresses, object.otherNodeAddresses)
					})
				if 0 >= len(addresses) {
					continue
				}
				var wg sync.WaitGroup
				wg.Add(len(object.otherNodeAddresses))
				for _, address := range object.otherNodeAddresses {
					apiURL := fmt.Sprintf("http://%s%s", address, "/api/v1/vote")
					routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
						defer wg.Done()
						values := url.Values{}
						values.Add("from", object.myAddress)
						res, err := http.PostForm(apiURL, values)
						if nil != err {
							glog.Error(err)
							return nil
						}
						if http.StatusOK != res.StatusCode {
							glog.Error("status error", res.Status)
							return nil
						}
						return nil
					}, "RaftVote")
				}
				wg.Wait()
			case NodeStateLeader:
				var addresses []string
				object.WithLock(false,
					func() {
						object.counter = 0
						if 0 >= len(object.followerAddresses) {
							return
						}
						addresses = make([]string, len(object.followerAddresses))
						copy(addresses, object.followerAddresses)
					})
				if 0 >= len(addresses) {
					continue
				}
				var wg sync.WaitGroup
				wg.Add(len(object.followerAddresses))
				for _, address := range object.followerAddresses {
					apiURL := fmt.Sprintf("http://%s%s", address, "/api/v1/heartbeat")
					routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
						defer wg.Done()
						values := url.Values{}
						values.Add("from", object.myAddress)
						values.Add("term", xint64.Int64(object.term).ToString())
						res, err := http.PostForm(apiURL, values)
						if nil != err {
							glog.Error(err)
							return nil
						}
						if http.StatusOK != res.StatusCode {
							glog.Error("status error", res.Status)
							return nil
						}
						var R struct {
							Leader bool   `json:"leader" form:"leader" binding:"required"`
							From   string `json:"from" form:"from" binding:"required"`
							Term   int64  `json:"term" form:"term" binding:"term"`
						}
						var raw []byte
						raw, err = ioutil.ReadAll(res.Body)
						if nil != err {
							glog.Error(err)
							return nil
						}
						if err = json.Unmarshal(raw, &R); nil != err {
							glog.Error(err)
							return nil
						}
						object.WithLock(false, func() {
							if NodeStateLeader != object.NodeState {
								return
							}
							if R.Leader && R.Term > object.term {
								object.counter = 0
								object.term = R.Term
								object.leaderAddress = R.From
								object.followerAddresses = make([]string, 0)
								object.NodeState = NodeStateFollower
							}
						})
						return nil
					}, "RaftVote")
				}
				wg.Wait()
			}
		}
	}
	object.wg.Done()
}

// Start 启动节点
func (object *Node) Start() {
	engine := gin.New()
	engine.Use(gin.Recovery())
	apiGroup := engine.Group("/api/v1")
	apiGroup.POST("/vote", func(ctx *gin.Context) {
		var req struct {
			From string `json:"from" form:"from" binding:"required"`
		}
		err := ctx.ShouldBind(&req)
		if nil != err {
			glog.Error(err)
			ctx.Status(http.StatusBadRequest)
			return
		}
		flag := false
		object.WithLock(false,
			func() {
				if 0 >= len(object.leaderAddress) {
					object.leaderAddress = req.From
					flag = true
					return
				}
			})
		if flag {
			values := url.Values{}
			values.Add("from", object.myAddress)
			var res *http.Response
			apiURL := fmt.Sprintf("http://%s%s", req.From, "/api/v1/vote/ack")
			res, err = http.PostForm(apiURL, values)
			if nil != err {
				glog.Error(err)
				flag = false
			} else if http.StatusOK != res.StatusCode {
				flag = false
				glog.Error("status error", res.Status)
			}
			object.WithLock(false,
				func() {
					if !flag {
						object.leaderAddress = ""
					} else {
						// 节点成为Follower
						object.NodeState = NodeStateFollower
					}
				})
		}
	})
	apiGroup.POST("/vote/ack", func(ctx *gin.Context) {
		var req struct {
			From string `json:"from" form:"from" binding:"required"`
		}
		err := ctx.ShouldBind(&req)
		if nil != err {
			glog.Error(err)
			ctx.Status(http.StatusBadRequest)
			return
		}
		object.WithLock(false,
			func() {
				object.followerAddresses = append(object.followerAddresses, req.From)
				if len(object.followerAddresses) > len(object.otherNodeAddresses)/2 {
					// 节点成为Leader
					object.term++ // 任期+1
					object.leaderAddress = object.myAddress
					object.NodeState = NodeStateLeader
				}
			})
	})
	apiGroup.POST("/heartbeat", func(ctx *gin.Context) {
		var req struct {
			From string `json:"from" form:"from" binding:"required"`
			Term int64  `json:"term" form:"term" binding:"required"`
		}
		err := ctx.ShouldBind(&req)
		if nil != err {
			glog.Error(err)
			ctx.Status(http.StatusBadRequest)
			return
		}
		object.WithLock(false,
			func() {
				flag := false
				switch object.NodeState {
				case NodeStateFollower, NodeStateCandidate:
					flag = true
				case NodeStateLeader:
					// Leader收到其它Leader发送的心跳消息&&其他Leader的任期大于自己
					if req.Term > object.term {
						flag = true
					}
				}
				type res struct {
					Leader bool   `json:"leader" form:"leader" binding:"required"`
					From   string `json:"from" form:"from" binding:"required"`
					Term   int64  `json:"term" form:"term" binding:"term"`
				}
				ctx.JSON(http.StatusOK, &res{
					Leader: NodeStateLeader == object.NodeState,
					From:   object.myAddress,
					Term:   object.term,
				})
				if flag {
					// 重置
					object.leaderAddress = req.From
					object.term = req.Term
					object.counter = 0
					object.followerAddresses = make([]string, 0)
					object.NodeState = NodeStateFollower
				}
			})
	})
	srv := &http.Server{
		Addr:    object.myAddress,
		Handler: engine,
	}
	routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
		if err := srv.ListenAndServe(); nil != err && http.ErrServerClosed != err {
			glog.Error(err)
		}
		return nil
	}, "RaftWebApi")
	routine_pool.GetInstance().PostTask(func(ctx context.Context, params []interface{}) interface{} {
		object.switchState()
		return nil
	}, "RaftSwitchState")
	event_bus.GetInstance().Mounting(reflect.TypeOf(&event.AppShutdownEvent{}),
		func(_ context.Context, param interface{}) {
			if priority_define.HTTPServiceShutdownPriority !=
				param.(*event.AppShutdownEvent).ShutdownPriority {
				return
			}
			ctx := context.Background()
			srv.Shutdown(ctx)
			select {
			case <-ctx.Done():
			}
			glog.Info("RaftWebApi done")
		})
}

// Stop 停止
func (object *Node) Stop() {
	object.cancel()
	object.wg.Wait()
}
