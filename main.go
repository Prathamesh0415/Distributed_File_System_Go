package main

import (
	//"fmt"
	"bytes"
	"fmt"
	//"io/ioutil"

	// "fmt"
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
		EncKey: newEncryptionKey(),
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
	time.Sleep(time.Second * 2)
	go s2.Start()
	time.Sleep(time.Second * 2)
	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("data_%d", i)
		data := bytes.NewReader([]byte("Some big data file"))
		s2.Store(key, data)
		time.Sleep(time.Millisecond * 500)
	

		s2.store.Delete(key)
		// // select{}
		r, err := s2.Get(key)
		if err != nil {
			fmt.Println(err)
		}
		buf := new(bytes.Buffer)
		io.Copy(buf, r)
		fmt.Println(string(buf.Bytes()))
		// select{}
		b, err := io.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	}
}



