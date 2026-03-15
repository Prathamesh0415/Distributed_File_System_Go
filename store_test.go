package main

import (
	"testing"

	//"github.com/stretchr/testify/assert"

	"bytes"
)

func TestStore(t *testing.T){
	opts := StoreOpts{
		PathTransformFunc: DefaultPathTransformFunc,
	}

	s := NewStore(opts)
	stream := bytes.NewReader([]byte("From jpg file"))
	err := s.writeStream("Hello", stream)
	if err != nil {
		t.Error(err)
	}
}
