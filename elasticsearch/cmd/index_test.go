package cmd

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/intelligentfish/gogo/elasticsearch/restClient"
)

func randomPath() string {
	raw := make([]byte, 4*8)
	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint64(raw[i*8:], rand.Uint64())
	}
	h := md5.New()
	h.Write(raw)
	return fmt.Sprintf("/%x", h.Sum(nil))
}

func TestIndex(t *testing.T) {
	type User struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	user := &User{Id: 1, Name: "neo.2", Age: 31}
	raw, err := json.Marshal(user)
	if nil != err {
		panic(err)
	}
	t.Log(string(raw))
	cmd := NewIndex(fmt.Sprintf("/users/user/%d", user.Id), raw)
	err = restClient.NewRestClient().Do(cmd, 5*time.Second)
	if nil != err {
		t.Error(err)
		return
	}
	t.Log(cmd)
}
