package main

import (
	"fmt"
	"log"
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

func (f *FileServer) loop() {
	defer func(){
		log.Println("File server stopped due to user quit action")
		f.Transport.Close()	
	}()
	select {
		case msg := <- f.Transport.Consume():
			fmt.Println(msg)
		case <- f.quitch:
			return 
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


func (f *FileServer) Start() error {
	if err := f.Transport.ListenAndAccept(); err != nil {
		return err
	}

	f.bootstrapNetwork()

	f.loop()

	return nil
}