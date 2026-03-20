package p2p

import "net"

// Represents a remote node
type Peer interface{
	RemoteAddr() net.Addr
	Close() error
	send([]byte) error
}

// handles connection between nodes
// can be tcp udp etc
type Transport interface {
	Dial(string) error
	ListenAndAccept() error	
	Consume() <-chan RPC
	Close() error
}