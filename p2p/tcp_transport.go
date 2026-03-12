package p2p

import (
	"fmt"
	"net"
	"sync"
)

type TCPPeer struct{
	conn net.Conn
	outbound bool
}

type TCPTransport struct {
	listenAddress string
	listener net.Listener

	mu sync.Mutex
	peers map[string]Peer
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer{
	return &TCPPeer{
		conn: conn,
		outbound: outbound,
	}
}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddress: listenAddr,
	}
}

func (t *TCPTransport) ListenAndAccept() error{
	var err error

	t.listener, err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	} 

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for{
		conn, err := t.listener.Accept()

		if err != nil {
			fmt.Printf("Error at start accept loop function: %s\n", err)
		}

		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)
	fmt.Printf("new connection at %v\n", peer)
}

