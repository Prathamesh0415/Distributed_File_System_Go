package main

import (
	//"fmt"
	"log"
	"time"

	"github.com/Prathamesh0415/fileserver/p2p"
) 

func OnPeer(peer p2p.Peer) error {
	peer.Close()
	return nil
}

func main(){
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddress: ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder: p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot: "3000_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport: tcpTransport,
	}
	FileServer := NewFileServer(fileServerOpts)

	go func() {
		time.Sleep(time.Second * 3)
		FileServer.Stop()
	}()

	if err := FileServer.Start(); err != nil {
		log.Fatal(err)
	}

}

