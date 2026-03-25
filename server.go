package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"time"

	//"strings"
	"sync"

	//"structs"

	"github.com/Prathamesh0415/fileserver/p2p"
)

type FileServerOpts struct {
	ID				  string
	EncKey			  []byte
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers map[string]p2p.Peer

	store *Store
	quitch chan struct{}
}

type Message struct {
	//From string
	Payload any
}

type MessageStoreFile struct {
	ID string
	Key string
	Size int64
}

type MessageGetFile struct {
	ID string
	Key string
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}

	if len(opts.ID) == 0 {
		opts.ID = generateId()
	}

	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch: 		make(chan struct{}),
		peers: 			make(map[string]p2p.Peer),
	}
}



func (f *FileServer) MessageGetFile() error {
	return nil
}

func (f *FileServer) Get(key string) (io.Reader, error) {
	if f.store.Has(f.ID, key) {
		fmt.Printf("[%s] Serving file (%s) from the local disk: \n", f.Transport.Addr(), key)
		_, r, err := f.store.Read(f.ID, key)
		return r, err
	}

	fmt.Printf("[%s] File (%s) not present locally connecting to network: \n", f.Transport.Addr(), key)

	msg := Message{
		Payload: MessageGetFile{
			ID: f.ID,
			Key: hashKey(key),
		},
	}

	if err := f.broadcast(&msg); err != nil {
		//f.peer.Send([]byte{p2p.IncomingMessage})
		return nil, err
	}

	time.Sleep(time.Millisecond * 500)

	for _, peer := range f.peers {
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		n, err := f.store.WriteDecrypt(f.EncKey, f.ID, key, io.LimitReader(peer, fileSize))
		//fileBuffer := new(bytes.Buffer)
		//fmt.Print("hello 1231")
		//n, err := io.CopyN(fileBuffer, peer, 18)
		//fmt.Print("hello")
		if err != nil {
			return nil, err
		}
		fmt.Printf("[%s] received %d bytes over the network from [%s]: \n",  f.Transport.Addr(), n, peer.RemoteAddr())
		//fmt.Println(string(fileBuffer.Bytes()))
		peer.CloseStream()
	}
	//select {}

	_, r, err := f.store.Read(f.ID, key)
	return r, err
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
	for{
		select {
			case rpc := <- f.Transport.Consume():
				var msg Message
				if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
					log.Println("decoding error: ", err)
				}
				if err := f.handleMessage(rpc.From, &msg); err != nil {
					log.Println("handle message error: ", err)
				}

				//time.Sleep(time.Second * 3)

				// fmt.Printf("received: %+v\n", msg)
				// peer, ok := f.peers[rpc.From]
				// if !ok {
				// 	fmt.Print("peer not found in peer map")
				// }
				// b := make([]byte, 1000)
				// if _, err := peer.Read(b); err != nil {
				// 	panic(err)
				// }
				// fmt.Printf("received: %s\n", string(b))

				// peer.(*p2p.TCPPeer).Wg.Done()
			case <- f.quitch:
				return 
		}
	}
}

func (f *FileServer) handleMessage(from string,msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return f.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return f.handleMessageGetFile(from, v)
		//fmt.Print("Getting done")
		//return err
	}
	return nil
}

func (f *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := f.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not present in peer map", from)
	}
	//tcpPeer := peer.(*p2p.TCPPeer)
	//fmt.Println("Hello1")
	//time.Sleep(time.Millisecond * 3)
	//print("hello")
	//<-tcpPeer.StreamReady
	n, err := f.store.Write(msg.ID, msg.Key, io.LimitReader(peer, msg.Size)) 
	if err != nil {
		return err
	}
	//fmt.Println("Hello2")
	fmt.Printf("%d bytes written to disk \n", n)
	peer.CloseStream()
	return nil 
}

func (f *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !f.store.Has(msg.ID, msg.Key) {
		return fmt.Errorf("[%s] Cannot read file %s from disk\n", f.Transport.Addr(), msg.Key)
	}
	fmt.Printf("[%s] Serving file (%s) over the network\n", f.Transport.Addr(), msg.Key)
	fileSize, r, err := f.store.Read(msg.ID, msg.Key)
	if err != nil {
		return err
	}

	rc, ok := r.(io.ReadCloser)
	if ok {
		fmt.Println("Closing readCloser")
		defer rc.Close()
	}

	peer, ok := f.peers[from]
	if !ok {
		return fmt.Errorf("peer %s is not present in the map\n", from)
	}
	//fmt.Print(peer.RemoteAddr())
	peer.Send([]byte{p2p.IncomingStream})
	//var fileSize int64 = 22
	binary.Write(peer, binary.LittleEndian, fileSize) 
	//time.Sleep(time.Second * 3)
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}
	fmt.Printf("[%s] written %d bytes over the network to %s\n", f.Transport.Addr(), n, from)
	//peer.(*p2p.TCPPeer).Wg.Done()
	return nil
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



func (f *FileServer) stream(p *Message) error {
	peers := []io.Writer{}
	for _, peer := range f.peers {
		peers = append(peers, peer)
	}
	mu := io.MultiWriter(peers...)
	return gob.NewEncoder(mu).Encode(p)
}

func (f *FileServer) broadcast(msg *Message) error {
	MsgBuf := new(bytes.Buffer)
	
	if err := gob.NewEncoder(MsgBuf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range f.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(MsgBuf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

func (f *FileServer) Store(key string, r io.Reader) error {
	FileBuffer := new(bytes.Buffer)
	tee := io.TeeReader(r, FileBuffer)	
	n, err := f.store.Write(f.ID, key, tee)
	if err != nil {
		return err
	}
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

	// fmt.Println("Hello")
	msg := Message{
		Payload: MessageStoreFile{
			ID: f.ID,
			Key: hashKey(key),
			Size: n + 16,
		},
	}

	if err := f.broadcast(&msg); err != nil {
		return err
	}	


	peers := []io.Writer{}

	for _, peer := range f.peers {
		peers = append(peers, peer)
	}

	time.Sleep(time.Millisecond * 5)

	mw := io.MultiWriter(peers...)
	mw.Write([]byte{p2p.IncomingStream})
	_, err = CopyEncrypt(f.EncKey, FileBuffer, mw)
	if err != nil {
		return err
	}



	//fmt.Println("Hello")

	

	//payload := []byte("THIS LARGE FILE")

	// fileBytes := FileBuffer.Bytes()

	// for _, peer := range f.peers {
	// 	peer.Send([]byte{p2p.IncomingStream})
	// 	//fmt.Println(FileBuffer.Bytes())
	// 	n, err := CopyEncrypt(f.EncKey, FileBuffer, peer)
	// 	// n, err := io.Copy(peer, bytes.NewReader(fileBytes))
	// 	// //fmt.Println(peer)
	// 	// if err != nil {
	// 	// 	return err
	// 	// }
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fmt.Printf("%d bytes copied to disk\n", n)
	// }

	//select{}

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
	gob.Register(MessageGetFile{})
}