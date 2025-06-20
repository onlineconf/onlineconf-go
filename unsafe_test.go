package onlineconf

import (
	"testing"
	"unsafe"
)

func TestSameAddress(t *testing.T) {
	orig := []byte{'a', 's', 'd', 'f'}
	str := b2s(orig)
	bytes := s2b(str)

	origAddr := *(*uintptr)(unsafe.Pointer(&orig))
	strAddr := *(*uintptr)(unsafe.Pointer(&str))
	bytesAddr := *(*uintptr)(unsafe.Pointer(&bytes))

	if strAddr != origAddr {
		t.Errorf("string data pointer mismatch: b2s(%#v) = %x, want %x", orig, strAddr, origAddr)
	}

	if bytesAddr != origAddr {
		t.Errorf("byte slice pointer mismatch: s2b(%q) = %x, want %x", str, bytesAddr, origAddr)
	}
}
