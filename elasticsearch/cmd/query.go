package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Query struct {
	path    string      `json:"path"`
	Index   string      `json:"_index"`
	Type    string      `json:"_type"`
	ID      string      `json:"_id"`
	Version int         `json:"_version"`
	Found   bool        `json:"found"`
	Source  interface{} `json:"_source"`
}

func (object *Query) HTTPMethod() string {
	return http.MethodGet
}
func (object *Query) Uri() string {
	return fmt.Sprintf("%s?pretty", object.path)
}
func (object *Query) SetRequestBody(w io.Writer) (err error) {
	return
}
func (object *Query) ProcessResponseBody(body []byte) (err error) {
	err = json.Unmarshal(body, object)
	return
}
func (object *Query) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
func (object *Query) SetSource(source interface{}) *Query {
	object.Source = source
	return object
}
func NewQuery(path string) *Query {
	return &Query{
		path:   path,
		Source: make(map[string]interface{}),
	}
}
