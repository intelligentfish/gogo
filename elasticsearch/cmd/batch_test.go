package cmd

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/intelligentfish/gogo/elasticsearch/restClient"
)

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (object *User) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}

func TestBatchIndex(t *testing.T) {
	var err error
	rc := restClient.NewRestClient()
	err = rc.Do(NewDelete("/users"), 5*time.Second)
	if nil != err {
		t.Error(err)
		return
	}
	var data []BatchOp
	for i := 1; i <= 31; i++ {
		raw, _ := json.Marshal(&User{
			Id:   i,
			Name: fmt.Sprintf("neo.%d", i),
			Age:  i,
		})
		data = append(data, &BatchIndex{
			Id:   i,
			Data: raw,
		})
	}
	cmd := NewBatch("/users/user", data...)
	err = rc.Do(cmd, 5*time.Second)
	if nil != err {
		t.Error(err)
		return
	}
	t.Log(cmd)
}
