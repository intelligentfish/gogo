package cmd

import (
	"bytes"
	"reflect"
	"strings"
)

type LineParser struct {
	body []byte
}

func (object *LineParser) Fill(ptr interface{}) {
	if 0 < len(object.body) {
		var key string
		var value string
		object.body = bytes.TrimSpace(object.body)
		index := bytes.Index(object.body, []byte("\n"))
		if -1 != index {
			key = string(bytes.TrimSpace(object.body[:index+1]))
			value = string(bytes.TrimSpace(object.body[index+1:]))
		} else {
			return
		}
		keys := strings.Split(key, " ")
		for i := len(keys) - 1; i >= 0; i-- {
			if "" == keys[i] {
				keys = append(keys[0:i], keys[i+1:]...)
			}
		}
		var values []string
		if -1 != index {
			values = strings.Split(value, " ")
			for i := len(values) - 1; i >= 0; i-- {
				if "" == values[i] {
					values = append(values[0:i], values[i+1:]...)
				}
			}
		} else {
			values = make([]string, len(keys))
		}
		m := make(map[string]string, len(keys))
		for i := 0; i < len(keys); i++ {
			m[keys[i]] = values[i]
		}
		rv := reflect.ValueOf(ptr)
		if rv.Kind() != reflect.Ptr || rv.IsNil() {
			return
		}
		rt := reflect.TypeOf(ptr)
		for i := 0; i < rt.Elem().NumField(); i++ {
			key := rt.Elem().Field(i).Tag.Get("json")
			rv.Elem().Field(i).SetString(m[key])
		}
	}
}

func NewLineParser(body []byte) *LineParser {
	return &LineParser{body: body}
}
