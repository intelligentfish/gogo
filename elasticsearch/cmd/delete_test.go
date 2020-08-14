package cmd

import (
	"testing"
	"time"

	"github.com/intelligentfish/gogo/elasticsearch/restClient"
)

func TestDelete(t *testing.T) {
	cmd := NewDelete("/users")
	err := restClient.NewRestClient().Do(cmd, 5*time.Second)
	if nil != err {
		t.Error(err)
		return
	}
	t.Log(cmd)
}
