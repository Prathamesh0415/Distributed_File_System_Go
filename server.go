package main

import (
	"fmt"
	"log"
	//"structs"

	"github.com/Prathamesh0415/fileserver/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
}

type FileServer struct {
	FileServerOpts
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
	}
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


func (f *FileServer) Start() error {
	if err := f.Transport.ListenAndAccept(); err != nil {
		return err
	}

	f.loop()

	return nil
}