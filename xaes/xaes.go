package xaes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// AES AES对象
type AES struct {
	key []byte
	iv  []byte
}

// NewAES 工厂方法
func NewAES(key []byte) *AES {
	return &AES{key: key}
}

// pkcs7Padding pkcs7填充
func (object *AES) pkcs7Padding(in []byte, blockSize int) []byte {
	if 0 == len(in)%blockSize {
		return in
	}
	padding := blockSize - len(in)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(in, padText...)
}

// pkcs7UnPadding pkcs7去掉填充
func (object *AES) pkcs7UnPadding(in []byte) []byte {
	length := len(in)
	if 0 >= length {
		return in
	}
	padding := in[length-1]
	start := length - 1
	end := length - int(padding)
	for start > end {
		if in[start] != padding {
			return in
		}
		start--
	}
	return in[:length-int(padding)]
}

// SetIV 设置IV向量
func (object *AES) SetIV(iv []byte) *AES {
	object.iv = iv
	return object
}

// Encrypt 加密
func (object *AES) Encrypt(in []byte) (out []byte, err error) {
	var block cipher.Block
	block, err = aes.NewCipher(object.key)
	if nil != err {
		return
	}
	blockSize := block.BlockSize()
	in = object.pkcs7Padding(in, blockSize)
	if nil == object.iv {
		object.iv = object.key[:blockSize]
	}
	blockMode := cipher.NewCBCEncrypter(block, object.iv)
	out = make([]byte, len(in))
	blockMode.CryptBlocks(out, in)
	return
}

// Decrypt 解密
func (object *AES) Decrypt(in []byte) (out []byte, err error) {
	var block cipher.Block
	block, err = aes.NewCipher(object.key)
	if nil != err {
		return
	}
	blockSize := block.BlockSize()
	if nil == object.iv {
		object.iv = object.key[:blockSize]
	}
	blockMode := cipher.NewCBCDecrypter(block, object.iv)
	out = make([]byte, len(in))
	blockMode.CryptBlocks(out, in)
	out = object.pkcs7UnPadding(out)
	return
}
