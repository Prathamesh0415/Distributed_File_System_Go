package p2p

// Represents a remote node
type Peer interface{
	Close() error
}

// handles connection between nodes
// can be tcp udp etc
type Transport interface {
	ListenAndAccept() error	
	Consume() <-chan RPC
}