package p2p

import (
	"encoding/gob"
	//"fmt"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *Message) error
}

type GOBDecoder struct{}

type DefaultDecoder struct{}

func (g GOBDecoder) Decode(r io.Reader, msg *Message) error {
	return gob.NewDecoder(r).Decode(msg)
}

func (g DefaultDecoder) Decode(r io.Reader, msg *Message) error {
	buf := make([]byte, 1028);
	n, err := r.Read(buf);
	if err != nil {
		return err
	}
	msg.Payload = buf[:n]
	return nil
}

