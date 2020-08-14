package cmd

import (
	"testing"
	"time"

	"github.com/intelligentfish/gogo/elasticsearch/restClient"
)

func TestHealth(t *testing.T) {
	health := &Health{}
	err := restClient.NewRestClient().Do(health, time.Second)
	if nil != err {
		t.Error(err)
		return
	}
	t.Log(health)
}
