package cmd

import (
	"encoding/json"
	"io"
	"net/http"
)

type Health struct {
	Epoch               string `json:"epoch"`
	Timestamp           string `json:"timestamp"`
	Cluster             string `json:"cluster"`
	Status              string `json:"status"`
	NodeTotal           string `json:"node.total"`
	NodeData            string `json:"node.data"`
	Shards              string `json:"shards"`
	Pri                 string `json:"pri"`
	Relo                string `json:"relo"`
	Init                string `json:"init"`
	Unassign            string `json:"unassign"`
	PendingTasks        string `json:"pending_tasks"`
	MaxTaskWaitTime     string `json:"max_task_wait_time"`
	ActiveShardsPercent string `json:"active_shards_percent"`
}

func (object *Health) HTTPMethod() string {
	return http.MethodGet
}
func (object *Health) Uri() string {
	return "/_cat/health?v"
}
func (object *Health) SetRequestBody(w io.Writer) (err error) { return }
func (object *Health) ProcessResponseBody(body []byte) (err error) {
	NewLineParser(body).Fill(object)
	return
}
func (object *Health) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
