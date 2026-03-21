package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	//"time"

	//"strings"
	"sync"

	//"structs"

	"github.com/Prathamesh0415/fileserver/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers map[string]p2p.Peer

	Store *Store
	quitch chan struct{}
}

type Message struct {
	//From string
	Payload any
}

type MessageStoreFile struct {
	Key string
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}

	return &FileServer{
		FileServerOpts: opts,
		Store:          NewStore(storeOpts),
		quitch: 		make(chan struct{}),
		peers: 			make(map[string]p2p.Peer),
	}
}

func (f *FileServer) OnPeer(p p2p.Peer) error {
	f.peerLock.Lock()
	defer f.peerLock.Unlock()
	f.peers[p.RemoteAddr().String()] = p

	log.Printf("Connected with remote: %s", p.RemoteAddr().String())

	return nil

}

func (f *FileServer) Stop() {
	close(f.quitch)
}

// func (f *FileServer) handleMessage(msg *Message) error {
// 	switch v := msg.Payload.(type) {
// 	case *DataMessage:
// 		fmt.Printf("received data %v", v)
// 	}
// 	return nil
// }

func (f *FileServer) loop() {
	defer func(){
		log.Println("File server stopped due to user quit action")
		f.Transport.Close()	
	}()
	for{
		select {
			case rpc := <- f.Transport.Consume():
				var msg Message
				if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
					log.Println(err)
				}
				fmt.Printf("received: %+v\n", msg)
				peer, ok := f.peers[rpc.From]
				if !ok {
					fmt.Print("peer not found in peer map")
				}
				b := make([]byte, 1000)
				if _, err := peer.Read(b); err != nil {
					panic(err)
				}
				fmt.Printf("received: %s\n", string(b))

				peer.(*p2p.TCPPeer).Wg.Done()
			case <- f.quitch:
				return 
		}
	}
}

func (f *FileServer) bootstrapNetwork() error {
	for _, addr := range f.BootstrapNodes {
		if len(addr) == 0 {continue}
		go func(addr string) {
			fmt.Println("attempting to connect with network: ", addr)
		
			if err := f.Transport.Dial(addr); err != nil {
				log.Println("Dial error: ", err)
			
			}
		}(addr)
		
	}
	return nil
}



func (f *FileServer) broadcast(p *Message) error {
	peers := []io.Writer{}
	for _, peer := range f.peers {
		peers = append(peers, peer)
	}
	mu := io.MultiWriter(peers...)
	return gob.NewEncoder(mu).Encode(p)
}



func (f *FileServer) StoreData(key string, r io.Reader) error {
	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r, buf)	
	// if err := f.Store.Write(key, tee); err != nil {
	// 	return err
	// }
	// _, err := io.Copy(buf, r)
	// if err != nil {
	// 	return err
	// }
    
	// p := &DataMessage{
	// 	Key: key,
	// 	Data: buf.Bytes(),
	// }

	// //fmt.Println(buf)
	// return f.broadcast(&Message{
	// 	From: "todo",
	// 	Payload: p,
	// })

	buf := new(bytes.Buffer)
	msg := Message{
		Payload: MessageStoreFile{
			Key: key,
		},
	}

	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range f.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	// time.Sleep(time.Second * 3)

	// payload := []byte("THIS LARGE FILE")

	// for _, peer := range f.peers {
	// 	if err := peer.Send(payload); err != nil {
	// 		return err
	// 	}
	// }


	return nil
}


func (f *FileServer) Start() error {
	if err := f.Transport.ListenAndAccept(); err != nil {
		return err
	}

	f.bootstrapNetwork()

	f.loop()

	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
}