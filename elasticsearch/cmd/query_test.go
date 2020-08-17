package cmd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/intelligentfish/gogo/elasticsearch/restClient"
)

func TestQuery(t *testing.T) {
	const defaultTimeout = 5 * time.Second
	rc := restClient.NewRestClient()
	rc.Do(NewDelete("/users"), defaultTimeout)
	raw, _ := json.Marshal(&User{
		Id:   1,
		Name: "neo.wang.1",
		Age:  1,
	})
	index := NewIndex("/users/user/1", raw)
	err := rc.Do(index, defaultTimeout)
	if nil != err {
		t.Error(err)
		return
	}
	t.Log(index)
	var user User
	query := NewQuery("/users/user/1").SetSource(&user)
	err = rc.Do(query, defaultTimeout)
	if nil != err {
		t.Error(err)
		return
	}
	t.Log(user.String())
}
