package restClient

import "io"

type CMD interface {
	HTTPMethod() string                          // http method
	Uri() string                                 // uri
	SetRequestBody(w io.Writer) (err error)      // set request body data
	ProcessResponseBody(body []byte) (err error) // process response data
}
