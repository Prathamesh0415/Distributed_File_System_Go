package p2p

import (
	"encoding/gob"
	//"fmt"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type GOBDecoder struct{}

type DefaultDecoder struct{}

func (g GOBDecoder) Decode(r io.Reader, msg *RPC) error {
	return gob.NewDecoder(r).Decode(msg)
}

func (g DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
	peekBuff := make([]byte, 1)
	if _, err := r.Read(peekBuff); err != nil {
		return err
	} 	
	//fmt.Print(string(peekBuff))
	
	stream := peekBuff[0] == IncomingStream
	//fmt.Print(stream)

	if stream {
		msg.Stream = true
		return nil
	}

	buf := make([]byte, 1028);
	n, err := r.Read(buf);
	if err != nil {
		//fmt.Printf("Hello")
		return err
	}
	msg.Payload = buf[:n]
	return nil
}

