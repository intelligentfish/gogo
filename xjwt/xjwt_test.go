package xjwt

import "testing"

func TestJWT(t *testing.T) {
	type Header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	type Payload struct {
		Sub  string `json:"sub"`
		Name string `json:"name"`
		Iat  int    `json:"iat"`
	}
	jwt := NewJWT(&Header{
		Alg: "HS256",
		Typ: "JWT",
	}, &Payload{
		Sub:  "1234567890",
		Name: "John Doe",
		Iat:  1516239022,
	}, "12345678901234567890123456789012")
	if "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.0IF0DV0RA3zx0YhD31jk5O-QNUBlK2BvmaB9476Xg_s" !=
		jwt.String() {
		t.Error("")

		return
	}
	other, err := ToJWT("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.0IF0DV0RA3zx0YhD31jk5O-QNUBlK2BvmaB9476Xg_s",
		"12345678901234567890123456789012")
	if nil != err {
		t.Error(err)
		return
	}
	header := other.Header.(map[string]interface{})
	payload := other.Payload.(map[string]interface{})
	t.Log(header)
	t.Log(payload)
	if header["alg"] != "HS256" ||
		header["typ"] != "JWT" ||
		payload["sub"] != "1234567890" ||
		payload["name"] != "John Doe" ||
		payload["iat"] != float64(1516239022) {
		t.Error("")
		return
	}
}
