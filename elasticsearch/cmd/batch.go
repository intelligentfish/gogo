package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// BatchOp batch operator
type BatchOp interface {
	ToRequestLine() string // convert to request line
}

// BatchIndex batch index operator
type BatchIndex struct {
	Id   int             // id
	Data json.RawMessage // raw json
}

// ToRequestLine convert to request line
func (object *BatchIndex) ToRequestLine() string {
	return fmt.Sprintf("{\"index\":{\"_id\":\"%d\"}}\n%s",
		object.Id,
		string(object.Data))
}

// BatchUpdate batch update operator
type BatchUpdate struct {
	Id   int                    // id
	Data map[string]interface{} // key value pair
}

// ToRequestLine convert to request line
func (object *BatchUpdate) ToRequestLine() string {
	raw, err := json.Marshal(object.Data)
	if nil != err {
		panic(err)
	}
	return fmt.Sprintf("{\"update\":{\"_id\":\"%d\"}}\n{\"doc\":%s}",
		object.Id,
		string(raw))
}

// BatchDelete delete batch operator
type BatchDelete struct {
	Id int // id
}

// ToRequestLine convert to request line
func (object *BatchDelete) ToRequestLine() string {
	return fmt.Sprintf("{\"delete\":{\"_id\":\"%d\"}}", object.Id)
}

// Batch
type Batch struct {
	path   string    // request path
	list   []BatchOp // operator list
	Took   int       `json:"took"`
	Errors bool      `json:"errors"`
	Items  []struct {
		Index struct {
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
			Status      int `json:"status"`
		} `json:"index"`
	} `json:"items"`
}

// HTTPMethod http method
func (object *Batch) HTTPMethod() string {
	return http.MethodPost
}

// Uri request uri
func (object *Batch) Uri() string {
	if '/' != object.path[len(object.path)-1] {
		object.path += "/"
	}
	return fmt.Sprintf("%s_bulk?pretty", object.path)
}

// SetRequestBody set request body
func (object *Batch) SetRequestBody(w io.Writer) (err error) {
	if nil == object.list {
		return
	}
	for _, op := range object.list {
		if _, err = w.Write([]byte(op.ToRequestLine())); nil != err {
			return
		}
		w.Write([]byte("\n"))
	}
	return
}

// ProcessResponseBody process response body
func (object *Batch) ProcessResponseBody(body []byte) (err error) {
	err = json.Unmarshal(body, object)
	return
}

// String string description
func (object *Batch) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}

// NewBatch factory method
func NewBatch(path string, list ...BatchOp) *Batch {
	return &Batch{path: path, list: list}
}
