package xjwt

import (
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"errors"
	"hash"
	"strings"
)

var (
	ErrJWTFormat = errors.New("jwt format error")
	ErrJWTSign   = errors.New("jwt sign error")
)

// JWT
type JWT struct {
	Header  interface{}
	Payload interface{}
	Secret  string
	hmac    hash.Hash
}

// NewJWTHmac 实例化HMAC
func NewJWTHmac(secret string) hash.Hash {
	return hmac.New(func() hash.Hash {
		return crypto.SHA256.New()
	}, []byte(secret))
}

// JWTToURLSafe 转换为URL安全
func JWTToURLSafe(in string) (out string) {
	out = strings.ReplaceAll(strings.ReplaceAll(in, "/", "_"),
		"+", "-")
	return
}

// JWTToRaw 转换为原始类型
func JWTToRaw(in string) (out string) {
	out = strings.ReplaceAll(strings.ReplaceAll(in, "_", "/"),
		"-", "+")
	return
}

// 转换为JWT
func ToJWT(jwtStr, secret string) (jwt *JWT, err error) {
	arr := strings.Split(JWTToRaw(jwtStr), ".")
	if 3 != len(arr) {
		err = ErrJWTFormat
		return
	}
	tmp := &JWT{}
	h := NewJWTHmac(secret)
	unsigned := arr[0] + "." + arr[1]
	h.Write([]byte(unsigned))
	signed := h.Sum(nil)
	if base64.RawStdEncoding.EncodeToString(signed) != arr[2] {
		err = ErrJWTSign
		return
	}
	var headerBytes []byte
	if headerBytes, err = base64.RawStdEncoding.DecodeString(arr[0]); nil != err {
		return
	}
	if err = json.Unmarshal(headerBytes, &tmp.Header); nil != err {
		return
	}
	var payloadBytes []byte
	if payloadBytes, err = base64.RawStdEncoding.DecodeString(arr[1]); nil != err {
		return
	}
	if err = json.Unmarshal(payloadBytes, &tmp.Payload); nil != err {
		return
	}
	jwt = tmp
	return
}

// NewJWT 工厂方法
func NewJWT(header, payload interface{},
	secret string) *JWT {
	object := &JWT{
		Header:  header,
		Payload: payload,
		Secret:  secret,
	}
	object.hmac = NewJWTHmac(object.Secret)
	return object
}

// String 字符串
func (object *JWT) String() (str string) {
	headerJson, _ := json.Marshal(object.Header)
	if 0 >= len(headerJson) {
		return
	}
	payloadJson, _ := json.Marshal(object.Payload)
	if 0 >= len(payloadJson) {
		return
	}
	unsigned := base64.RawStdEncoding.EncodeToString(headerJson) +
		"." +
		base64.RawStdEncoding.EncodeToString(payloadJson)
	object.hmac.Reset()
	object.hmac.Write([]byte(unsigned))
	str = JWTToURLSafe(unsigned + "." + base64.RawStdEncoding.EncodeToString(object.hmac.Sum(nil)))
	return
}

// Verify 校验
func (object *JWT) Verify(str string) (err error) {
	arr := strings.Split(JWTToRaw(str), ".")
	if 3 != len(arr) {
		err = ErrJWTFormat
		return
	}
	unsigned := arr[0] + "." + arr[1]
	object.hmac.Reset()
	object.hmac.Write([]byte(unsigned))
	signed := object.hmac.Sum(nil)
	if base64.RawStdEncoding.EncodeToString(signed) != arr[2] {
		err = ErrJWTSign
		return
	}
	return
}
