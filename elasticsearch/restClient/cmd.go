package restClient

import "io"

// CMD reset client cmd
type CMD interface {
	HTTPMethod() string                          // http method
	Uri() string                                 // uri
	SetRequestBody(w io.Writer) (err error)      // set request body data
	ProcessResponseBody(body []byte) (err error) // process response data
}
