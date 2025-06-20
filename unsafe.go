//nolint:gosec,govet,revive
package onlineconf

import (
	"unsafe"
)

// gc-friendly version of [reflect.StringHeader]
type stringHeader struct {
	data *byte
	len  int
}

func s2b(s string) []byte {
	strHdr := *(*stringHeader)(unsafe.Pointer(&s))
	return unsafe.Slice(strHdr.data, strHdr.len)
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
