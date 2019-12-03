package gateway

import (
	"github.com/intelligentfish/gogo/app"
	"testing"
)

func TestGateway(t *testing.T) {
	NewGateway().AddUpstream(NewUpstream(UpstreamTypeTCP,
		10088,
		"127.0.0.1",
		80,
		nil,
		nil)).Start()
	app.GetInstance().WaitShutdown()
}
