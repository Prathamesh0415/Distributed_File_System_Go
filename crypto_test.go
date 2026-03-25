package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCopyEncrypt(t *testing.T) {
	payload := "Hello this is text to be encrypted"
	
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)
	key := newEncryptionKey()
	_, err := CopyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(dst.Bytes())

	out := new(bytes.Buffer)
	nw, err := CopyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)
	}
	if nw != 16 + len(payload) {
		t.Fail()
	}
	if out.String() != payload {
		t.Error("Encryption failed")	
	}
}