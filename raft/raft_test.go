package raft

import (
	"flag"
	"github.com/intelligentfish/gogo/app"
	"testing"
)

func TestRaft(t *testing.T) {
	flag.Parse()
	addresses := []string{"localhost:8080", "localhost:8081", "localhost:8082"}
	NewNode(addresses[0], addresses[1], addresses[2]).Start()
	NewNode(addresses[1], addresses[0], addresses[2]).Start()
	NewNode(addresses[2], addresses[0], addresses[1]).Start()
	app.GetInstance().WaitShutdown()
}
