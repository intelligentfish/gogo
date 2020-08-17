package cmd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/intelligentfish/gogo/elasticsearch/restClient"
)

func TestUpdate(t *testing.T) {
	const defaultTimeout = 5 * time.Second
	rc := restClient.NewRestClient()
	rc.Do(NewDelete("/users"), defaultTimeout)
	raw, _ := json.Marshal(&User{
		Id:   1,
		Name: "neo.wang.1",
		Age:  1,
	})
	index := NewIndex("/users/1", raw)
	err := rc.Do(index, defaultTimeout)
	if nil != err {
		t.Error(err)
		return
	}
	update := NewUpdate("/users/1").SetUpdates(map[string]interface{}{"age": 2})
	err = rc.Do(update, defaultTimeout)
	if nil != err {
		t.Error(err)
		return
	}
	t.Log(update)
}
