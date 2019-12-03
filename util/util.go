package util

import (
	"bytes"
	"crypto"
	"encoding/gob"
	"fmt"
	"github.com/fatih/structs"
	"github.com/intelligentfish/gogo/triple"
	"github.com/intelligentfish/gogo/tuple"
	"io"
	"net"
	"os"
	"path/filepath"
	"reflect"
)

// Hash
func Hash(hash crypto.Hash, in []byte) (out []byte) {
	h := hash.New()
	h.Write(in)
	return h.Sum(nil)
}

// MD5
func MD5(in []byte) (out []byte) {
	return Hash(crypto.MD5, in)
}

// MD5String
func MD5String(in []byte) (out string) {
	out = fmt.Sprintf("%x", MD5(in))
	return
}

// ReaderHash
func ReaderHash(hash crypto.Hash, reader io.Reader) (out []byte, err error) {
	h := hash.New()
	var n int
	chunk := make([]byte, 1<<12)
	for {
		n, err = reader.Read(chunk)
		if nil != err {
			if io.EOF == err {
				err = nil
				if 0 < n {
					h.Write(chunk[:n])
				}
				out = h.Sum(nil)
				return
			}
		}
		h.Write(chunk[:n])
	}
	return
}

// ReaderMD5
func ReaderMD5(reader io.Reader) (out []byte, err error) {
	return ReaderHash(crypto.MD5, reader)
}

// SHA256
func SHA256(in []byte) (out []byte) {
	return Hash(crypto.SHA256, in)
}

// ReaderSHA256
func ReaderSHA256(reader io.Reader) (out []byte, err error) {
	return ReaderHash(crypto.SHA256, reader)
}

// MyIPAddresses 获取本机IP地址
func MyIPAddresses() (ips []string) {
	if addrArr, err := net.InterfaceAddrs(); nil == err {
		for _, addr := range addrArr {
			if ipAddr, ok := addr.(*net.IPNet); ok && !ipAddr.IP.IsLoopback() {
				v4 := ipAddr.IP.To4()
				if nil != v4 {
					v4Str := v4.String()
					if 0 < len(v4Str) {
						ips = append(ips, v4.String())
					}
				}
			}
		}
	}
	return
}

// MyMACs 获取本机MAC地址
func MyMACs() (macs []string) {
	if netInterfaceAddr, err := net.Interfaces(); nil == err {
		for _, netInterface := range netInterfaceAddr {
			mac := netInterface.HardwareAddr.String()
			if 0 < len(mac) {
				macs = append(macs, mac)
			}
		}
	}
	return
}

// Combinations2 组合
func Combinations2(arr []int, callback func(tuple *tuple.Tuple)) {
	for i := 0; i < len(arr)-1; i++ {
		for j := i + 1; j < len(arr); j++ {
			callback(&tuple.Tuple{
				A: arr[i],
				B: arr[j],
			})
		}
	}
	return
}

// Combinations3 组合
func Combinations3(arr []int, callback func(triple *triple.Triple)) {
	for i := 0; i < len(arr)-2; i++ {
		for j := i + 1; j < len(arr)-1; j++ {
			for k := j + 1; k < len(arr); k++ {
				callback(&triple.Triple{
					A: arr[i],
					B: arr[j],
					C: arr[k],
				})
			}
		}
	}
	return
}

// Copy 复制
func Copy(dst, src interface{}) (err error) {
	sb := &bytes.Buffer{}
	enc := gob.NewEncoder(sb)
	if err = enc.Encode(src); nil != err {
		return
	}

	dec := gob.NewDecoder(sb)
	err = dec.Decode(dst)
	return
}

// Struct2Map 结构体转换为Map
func Struct2Map(st interface{}, tagStr string) map[string]interface{} {
	m := make(map[string]interface{})
	for _, v := range structs.Fields(st) {
		flag := false
		switch v.Kind() {
		case reflect.Ptr:
			if reflect.Struct == reflect.ValueOf(v).Elem().Kind() {
				flag = true
			}
		case reflect.Struct:
			flag = true
		}
		if !flag {
			m[v.Tag(tagStr)] = v.Value()
		} else {
			for k, v := range Struct2Map(v.Value(), tagStr) {
				m[k] = v
			}
		}
	}
	return m
}

// PanicOnError 错误时崩溃
func PanicOnError(err error) {
	if nil != err {
		panic(err)
	}
}

// EnsureDirExists 确保目录存在
func EnsureDirExists(dirname string, mode os.FileMode) (err error) {
	if !filepath.IsAbs(dirname) {
		if dirname, err = filepath.Abs(dirname); nil != err {
			return
		}
	}
	var fi os.FileInfo
	if fi, err = os.Stat(dirname); nil != err {
		if _, ok := err.(*os.PathError); !ok {
			return
		}
		err = nil
	}
	if nil != fi && !fi.IsDir() {
		var absPath string
		absPath, err = filepath.Abs(dirname)
		if err = os.Remove(absPath); nil != err {
			return
		}
		err = os.Mkdir(dirname, mode)
	}
	return
}
