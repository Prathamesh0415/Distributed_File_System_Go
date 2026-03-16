package main

import (
	"fmt"
	"io"
	//"io/ioutil"
	"testing"

	//"github.com/stretchr/testify/assert"

	"bytes"
)

func TestDelete(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "somepictures"
	data := []byte("Some jpg pictures");
	s.writeStream(key, bytes.NewReader(data))
	_, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}
	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}

func TestPathTransformFunc(t *testing.T){
	key := "mypictures"
	path := CASPathTransformFunc(key)
	expectedPath := "abac3/e6bbd/3e468/fb226/de721/27008/bab22/b8de4"
	if path.Pathname != expectedPath {
		t.Errorf("path %s should be equal to %s\n", path, expectedPath)
	}
}

func TestStore(t *testing.T){
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "somepictures"
	data := []byte("Some jpg pictures");
	s.writeStream(key, bytes.NewReader(data))
	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}
	if ok := s.Has(key); !ok {
		t.Errorf("expected to hace key %s", key)
	}
	buf, err := io.ReadAll(r)
	fmt.Println(string(buf))
	if string(data) != string(buf){
		t.Errorf("want %s have %s", data, buf)
	}
}
