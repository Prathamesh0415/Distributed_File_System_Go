package main

import (
	"fmt"
	"io"
	//"os"

	//"io/ioutil"
	"testing"

	//"github.com/stretchr/testify/assert"

	"bytes"
)

func TestPathTransformFunc(t *testing.T){
	key := "mypictures"
	path := CASPathTransformFunc(key)
	expectedPath := "abac3/e6bbd/3e468/fb226/de721/27008/bab22/b8de4"
	if path.Pathname != expectedPath {
		t.Errorf("path %s should be equal to %s\n", path, expectedPath)
	}
}

func TestStore(t *testing.T){
	s := newStore()
	defer tearDown(t, s)
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("foo_%d", i)
		data := []byte("Some jpg pictures");
		if _, err := s.writeStream(key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}

		if ok := s.Has(key); !ok {
			t.Errorf("expected to hace key %s", key)
		}

		_, r, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}
		
		buf, err := io.ReadAll(r)
		fmt.Println(string(buf))
		if string(data) != string(buf){
			t.Errorf("want %s have %s", data, buf)
		}

		if err := s.Delete(key); err != nil {
			t.Error(err)
		}

		if ok := s.Has(key); ok {
			t.Errorf("Expected not to have %s key \n", key)
		}
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func tearDown(t *testing.T, s *Store){
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}