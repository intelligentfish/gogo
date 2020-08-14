package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Index index op
// if request body is nil, check "NoBody" fields
// otherwise check others fields
type Index struct {
	path                     string `json:"path"`
	body                     []byte `json:"body"`
	NoBodyAcknowledged       bool   `json:"acknowledged"`
	NoBodyShardsAcknowledged bool   `json:"shards_acknowledged"`
	NoBodyIndex              string `json:"index"`
	Index                    string `json:"_index"`
	Type                     string `json:"_type"`
	ID                       string `json:"_id"`
	Version                  int    `json:"_version"`
	Result                   string `json:"result"`
	Shards                   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	SeqNo       int `json:"_seq_no"`
	PrimaryTerm int `json:"_primary_term"`
}

func (object *Index) HTTPMethod() string {
	return http.MethodPost
}
func (object *Index) Uri() string {
	return fmt.Sprintf("%s?pretty", object.path)
}
func (object *Index) SetRequestBody(w io.Writer) (err error) {
	if nil != object.body {
		_, err = w.Write(object.body)
	}
	return
}
func (object *Index) ProcessResponseBody(body []byte) (err error) {
	err = json.Unmarshal(body, object)
	return
}
func (object *Index) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
func NewIndex(body string, data []byte) *Index {
	return &Index{path: body, body: data}
}
