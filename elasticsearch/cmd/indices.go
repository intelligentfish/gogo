package cmd

import (
	"encoding/json"
	"io"
	"net/http"
)

type Indices struct {
	Health       string `json:"health"`
	Status       string `json:"status"`
	Index        string `json:"index"`
	Uuid         string `json:"uuid"`
	Pri          string `json:"pri"`
	Rep          string `json:"rep"`
	DocsCount    string `json:"docs.count"`
	DocsDeleted  string `json:"docs.deleted"`
	StoreSize    string `json:"store.size"`
	PriStoreSize string `json:"pri.store.size"`
}

func (object *Indices) HTTPMethod() string {
	return http.MethodGet
}
func (object *Indices) Uri() string {
	return "/_cat/indices?v"
}
func (object *Indices) SetRequestBody(w io.Writer) (err error) { return }
func (object *Indices) ProcessResponseBody(body []byte) (err error) {
	NewLineParser(body).Fill(object)
	return
}
func (object *Indices) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}
