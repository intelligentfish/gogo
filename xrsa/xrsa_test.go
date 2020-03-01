package xrsa

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func Test(t *testing.T) {
	rsaObj := &RSA{}
	err := rsaObj.MakeKey(2048)
	if nil != err {
		t.Error(err)
		return
	}
	privateKey := rsaObj.PrivateKey
	publicKey := &rsaObj.PrivateKey.PublicKey
	var raw []byte
	if raw, err = ioutil.ReadFile("xrsa.go"); nil != err {
		t.Error(err)
		return
	}
	var encrypted []byte
	if encrypted, err = EncryptWithPublicKey(publicKey, raw); nil != err {
		t.Error(err)
		return
	}
	var decrypted []byte
	if decrypted, err = DecryptWithPrivateKey(privateKey, encrypted); nil != err {
		t.Error(err)
		return
	}
	if !bytes.Equal(raw, decrypted) {
		t.Error("bytes.Equal(raw, decrypted)")
		return
	}
}
