package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Update struct {
	path    string
	script  string
	updates map[string]interface{}
	Index   string `json:"_index"`
	Type    string `json:"_type"`
	ID      string `json:"_id"`
	Version int    `json:"_version"`
	Result  string `json:"result"`
	Shards  struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	SeqNo       int `json:"_seq_no"`
	PrimaryTerm int `json:"_primary_term"`
}

func (object *Update) HTTPMethod() string {
	return http.MethodPost
}
func (object *Update) Uri() string {
	if '/' != object.path[len(object.path)-1] {
		object.path += "/"
	}
	return fmt.Sprintf("%s_update?pretty", object.path)
}
func (object *Update) SetRequestBody(w io.Writer) (err error) {
	if nil != object.updates && 0 < len(object.updates) {
		type update struct {
			Doc []byte `json:"doc"`
		}
		var doc []byte
		doc, err = json.Marshal(object.updates)
		if nil != err {
			return
		}
		doc, err = json.Marshal(&update{Doc: doc})
		if nil == err {
			w.Write(doc)
		}
	} else if 0 < len(object.script) {
		w.Write([]byte(fmt.Sprintf(`{"script":"%s"}`, object.script)))
	}
	return
}
func (object *Update) ProcessResponseBody(body []byte) (err error) {
	err = json.Unmarshal(body, object)
	return
}
func (object *Update) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
func (object *Update) SetUpdates(updates map[string]interface{}) *Update {
	object.updates = updates
	return object
}
func (object *Update) SetScript(script string) *Update {
	object.script = script
	return object
}
func NewUpdate(path string) *Update {
	return &Update{path: path}
}
