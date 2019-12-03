package request_params

import (
	"bytes"
	"fmt"
	"github.com/intelligentfish/gogo/util"
	"sort"
)

// 请求参数
type RequestParams map[string]interface{}

// 计算请求参数签名
func (object RequestParams) Sign(key string) string {
	var keys []string
	for k := range object {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sb := &bytes.Buffer{}
	for _, k := range keys {
		fmt.Fprintf(sb, `%s=%s&`, k, fmt.Sprint(object[k]))
	}
	fmt.Fprint(sb, key)
	return string(util.MD5(sb.Bytes()))
}
