package xrsa

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

// EncryptWithPublicKey 公钥加密
func EncryptWithPublicKey(publicKey *rsa.PublicKey, in []byte) (out []byte, err error) {
	var chunk []byte
	chunkSize := publicKey.N.BitLen()/8 - 11
	buf := &bytes.Buffer{}
	for 0 < len(in) {
		if chunkSize > len(in) {
			chunkSize = len(in)
		}
		if chunk, err = rsa.EncryptPKCS1v15(rand.Reader,
			publicKey,
			in[:chunkSize]); nil != err {
			return
		}
		buf.Write(chunk)
		in = in[chunkSize:]
	}
	out = buf.Bytes()
	return
}

// DecryptWithPrivateKey 私钥解密
func DecryptWithPrivateKey(privateKey *rsa.PrivateKey, in []byte) (out []byte, err error) {
	var chunk []byte
	chunkSize := privateKey.N.BitLen() / 8
	buf := &bytes.Buffer{}
	for 0 < len(in) {
		if chunkSize > len(in) {
			chunkSize = len(in)
		}
		if chunk, err = rsa.DecryptPKCS1v15(rand.Reader,
			privateKey,
			in[:chunkSize]); nil != err {
			return
		}
		buf.Write(chunk)
		in = in[chunkSize:]
	}
	out = buf.Bytes()
	return
}

// RSA
type RSA struct {
	PrivateKey *rsa.PrivateKey
}

// MakeKey 创建Key
func (object *RSA) MakeKey(keySize int) (err error) {
	object.PrivateKey, err = rsa.GenerateKey(rand.Reader, keySize)
	return
}

// PrivateToPem 私钥转换为PEM格式
func (object *RSA) PrivateToPem() (str string, err error) {
	var bytes []byte
	bytes, err = x509.MarshalPKCS8PrivateKey(object.PrivateKey)
	if nil != err {
		return
	}
	str = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: bytes,
	}))
	return
}

// PublicToPem 公钥转换为PEM格式
func (object *RSA) PublicToPem() (str string,
	err error) {
	var bytes []byte
	bytes, err = x509.MarshalPKIXPublicKey(&object.PrivateKey.PublicKey)
	if nil != err {
		return
	}
	str = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: bytes,
	}))
	return
}

// Encrypt 加密
func (object *RSA) Encrypt(in []byte) (out []byte, err error) {
	return EncryptWithPublicKey(&object.PrivateKey.PublicKey, in)
}

// Decrypt 解密
func (object *RSA) Decrypt(in []byte) (out []byte, err error) {
	return DecryptWithPrivateKey(object.PrivateKey, in)
}

// Sign 签名
func (object *RSA) Sign(hash crypto.Hash, src []byte) (sign []byte, err error) {
	h := hash.New()
	h.Write(src)
	sign, err = rsa.SignPKCS1v15(rand.Reader, object.PrivateKey, hash, h.Sum(nil))
	return
}

// Verify 验证签名
func (object *RSA) Verify(hash crypto.Hash, src, sign []byte) (err error) {
	h := hash.New()
	h.Write(src)
	err = rsa.VerifyPKCS1v15(&object.PrivateKey.PublicKey, hash, h.Sum(nil), sign)
	return
}
