package main

import (
	//"fmt"
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	//"time"

	"github.com/Prathamesh0415/fileserver/p2p"
) 


func makeServer(listenAddr string, nodes ...string) *FileServer{
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder: p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot: listenAddr + "_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport: tcpTransport,
		BootstrapNodes: nodes,
	}

	f := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = f.OnPeer 
	return f
}

func main(){
	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")
	go func() {
		log.Fatal(s1.Start())	
	}()
	go s2.Start()
	time.Sleep(time.Second * 2)
	// data := bytes.NewReader([]byte("Some big data file"))
	// s2.Store("data", data)
	
	r, err :=s2.Get("data")
	if err != nil {
		fmt.Println(err)
	}
	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	fmt.Println(string(buf.Bytes()))
	select{}
}



