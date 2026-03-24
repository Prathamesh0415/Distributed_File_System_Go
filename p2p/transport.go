package p2p

import "net"

// Represents a remote node
type Peer interface{
	net.Conn
	Send([]byte) error
	CloseStream()
}

// handles connection between nodes
// can be tcp udp etc
type Transport interface {
	Addr() string
	Dial(string) error
	ListenAndAccept() error	
	Consume() <-chan RPC
	Close() error
	//ListenAddr() string
}