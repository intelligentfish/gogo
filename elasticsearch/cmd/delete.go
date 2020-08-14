package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Delete struct {
	path         string
	Acknowledged bool `json:"acknowledged"`
}

func (object *Delete) HTTPMethod() string {
	return http.MethodDelete
}
func (object *Delete) Uri() string {
	return fmt.Sprintf("%s?pretty", object.path)
}
func (object *Delete) SetRequestBody(w io.Writer) (err error) { return }
func (object *Delete) ProcessResponseBody(body []byte) (err error) {
	err = json.Unmarshal(body, object)
	return
}
func (object *Delete) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
func NewDelete(path string) *Delete {
	return &Delete{path: path}
}
