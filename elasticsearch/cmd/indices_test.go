package cmd

import (
	"testing"
	"time"

	"github.com/intelligentfish/gogo/elasticsearch/restClient"
)

func TestIndices(t *testing.T) {
	health := &Indices{}
	err := restClient.NewRestClient().Do(health, time.Second)
	if nil != err {
		t.Error(err)
		return
	}
	t.Log(health)
}
