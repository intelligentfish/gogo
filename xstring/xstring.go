package xstring

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strconv"
	"strings"
)

var (
	ErrPemParseError = errors.New("parse pem block error")
	ErrKeyTypeNotRSA = errors.New("key type is not rsa")
)

// String 扩展string
type String string

// ToInt 字符串转换为int型
func (object String) ToInt(panicWhenError bool) int {
	v, err := strconv.Atoi(string(object))
	if panicWhenError && nil != err {
		panic(err)
	}
	return v
}

// ToIntList 字符串转换为int数组
func (object String) ToIntList(sep string, panicWhenError bool) []int {
	var err error
	strValues := strings.Split(string(object), sep)
	intValues := make([]int, len(strValues))
	for index, v := range strValues {
		intValues[index], err = strconv.Atoi(v)
		if panicWhenError && nil != err {
			panic(err)
		}
	}
	return intValues
}

// IsEmpty 是否为空
func (object String) IsEmpty() bool {
	return 0 >= len(string(object))
}

// ToPrivateKey 转换为rsa.PrivateKey
func (object String) ToPrivateKey() (privateKey *rsa.PrivateKey,
	err error) {
	block, _ := pem.Decode([]byte(string(object)))
	if nil == block {
		err = ErrPemParseError
		return
	}
	var any interface{}
	any, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	if nil != err {
		return
	}
	switch anyObject := any.(type) {
	case *rsa.PrivateKey:
		privateKey = anyObject
	default:
		err = ErrKeyTypeNotRSA
	}
	return
}

// ToPublicKey 转换为rsa.PublicKey
func (object String) ToPublicKey() (publicKey *rsa.PublicKey,
	err error) {
	block, _ := pem.Decode([]byte(string(object)))
	if nil == block {
		err = ErrPemParseError
		return
	}
	var any interface{}
	any, err = x509.ParsePKIXPublicKey(block.Bytes)
	if nil != err {
		return
	}
	switch anyObject := any.(type) {
	case *rsa.PublicKey:
		publicKey = anyObject
	default:
		err = ErrKeyTypeNotRSA
	}
	return
}
