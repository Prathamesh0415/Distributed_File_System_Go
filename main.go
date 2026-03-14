package main

import (
	//"fmt"
	"log"

	"github.com/Prathamesh0415/fileserver/p2p"
) 

func main(){
	tr := p2p.NewTCPTransport(p2p.TCPTransportOpts{
		ListenAddress: ":4000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder: p2p.DefaultDecoder{},
	})

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal((err))
	}

	select {}
}

