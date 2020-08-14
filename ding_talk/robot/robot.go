package robot

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/intelligentfish/gogo/ding_talk"
)

var (
	robotOnce sync.Once // robot once
	robotInst *Robot    // robot singleton
)

// BaseMsg base message
type BaseMsg struct {
	MsgType string `json:"msgtype"` // message type
}

// BaseResp base response
type BaseResp struct {
	ErrorCode int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
}

// TxtMsgContent text message content
type TxtMsgContent struct {
	Content string `json:"content"`
}

// TxtMsgAt text message at
type TxtMsgAt struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool     `json:"isAtAll"`
}

// TxtMsg text message
type TxtMsg struct {
	BaseMsg
	Text TxtMsgContent `json:"text"`
	At   TxtMsgAt      `json:"at"`
}

// AsyncTask async task
type AsyncTask struct {
	content   string                            // content
	atMobiles []string                          // @ mobiles
	isAtAll   bool                              // whether @ all
	callback  func(result *BaseResp, err error) // async callback
}

// Robot Robot object
type Robot struct {
	stoppedFlag int32           // stopped flag
	wg          sync.WaitGroup  // wait group
	apiURL      string          // api url
	token       string          // token
	taskChCap   int             // async task chain capacity
	concurrency int             // async concurrency
	taskCh      chan *AsyncTask // async task chain
}

// Option option for robot
type Option func(object *Robot)

// ApiUrlOption api url option
func ApiUrlOption(apiURL string) Option {
	return func(object *Robot) {
		object.apiURL = apiURL
	}
}

// TokenOption token option
func TokenOption(token string) Option {
	return func(object *Robot) {
		object.token = token
	}
}

// TaskChainCapOption task chain capacity option
func TaskChainCapOption(cap int) Option {
	return func(object *Robot) {
		object.taskChCap = cap
	}
}

// ConcurrencyOption concurrency option
func ConcurrencyOption(num int) Option {
	return func(object *Robot) {
		object.concurrency = num
	}
}

// loop send text loop
func (object *Robot) loop() {
	defer object.wg.Done()
loop:
	for {
		select {
		case task, ok := <-object.taskCh:
			if !ok {
				break loop
			}
			task.callback(object.SendText(task.content, task.atMobiles, task.isAtAll))
		}
	}
}

// sign sign method
func (object *Robot) sign(timestamp int64) string {
	content := fmt.Sprintf("%d\n%s", timestamp, object.token)
	hmacObj := hmac.New(sha256.New, []byte(object.token))
	if _, err := hmacObj.Write([]byte(content)); nil != err {
		panic(err)
	}
	return url.QueryEscape(base64.StdEncoding.EncodeToString(hmacObj.Sum(nil)))
}

// SendText send text message
func (object *Robot) SendText(content string,
	atMobiles []string,
	isAtAll bool) (result *BaseResp, err error) {
	var raw []byte
	if raw, err = json.Marshal(&TxtMsg{
		BaseMsg: BaseMsg{MsgType: "text"},
		Text:    TxtMsgContent{Content: content},
		At:      TxtMsgAt{AtMobiles: atMobiles, IsAtAll: isAtAll},
	}); nil != err {
		panic(err)
	}
	now := time.Now().Unix() * 1000
	postURL := fmt.Sprintf("%s&timestamp=%d&sign=%s", object.apiURL, now, object.sign(now))
	var resp *http.Response
	resp, err = http.Post(postURL, "application/json", bytes.NewBuffer(raw))
	if nil != resp && nil != resp.Body {
		defer func() {
			if err = resp.Body.Close(); nil != err {
				fmt.Fprintln(os.Stderr, err)
			}
		}()
	}
	if nil != err {
		return
	}
	if nil != resp.Body {
		raw, _ = ioutil.ReadAll(resp.Body)
		var resObj BaseResp
		if err = json.Unmarshal(raw, &resObj); nil == err {
			result = &resObj
		}
	}
	return
}

// AsyncSendText async send text
func (object *Robot) AsyncSendText(content string,
	atMobiles []string,
	isAtAll bool,
	callback func(result *BaseResp, err error)) {
	select {
	case object.taskCh <- &AsyncTask{
		content:   content,
		atMobiles: atMobiles,
		isAtAll:   isAtAll,
		callback:  callback}:
	default:
		break
	}
}

// StopAndJoin stop robot and wait done
func (object *Robot) StopAndJoin() {
	if !atomic.CompareAndSwapInt32(&object.stoppedFlag, 0, 1) {
		return
	}
	close(object.taskCh)
	object.wg.Wait()
}

// NewRobot factory method
func NewRobot(options ...Option) *Robot {
	object := &Robot{}
	for _, opt := range options {
		opt(object)
	}
	if 0 >= object.taskChCap {
		object.taskChCap = 1024
	}
	if 0 >= object.concurrency {
		object.concurrency = 4
	}
	object.taskCh = make(chan *AsyncTask, object.taskChCap)
	object.wg.Add(object.concurrency)
	for i := 0; i < object.concurrency; i++ {
		go object.loop()
	}
	return object
}

// Inst singleton
func Inst() *Robot {
	robotOnce.Do(func() {
		robotInst = NewRobot(
			ApiUrlOption(ding_talk.WebHook),
			TokenOption(ding_talk.Token),
		)
	})
	return robotInst
}
