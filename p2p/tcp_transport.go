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

type TCPTransportOpts struct {
	ListenAddress string
	HandshakeFunc HandshakeFunc
	Decoder Decoder
}

type TCPTransport struct {
	TCPTransportOpts
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

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
	}
}

func (t *TCPTransport) ListenAndAccept() error{
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddress)
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

type temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)
	fmt.Printf("new connection at %v\n", peer)

	if err := t.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Printf("TCP Handshake error :%s\n", err)
		return
	}

	msg := &Message{};
	//buf := make([]byte, 1024)
	for {

		err := t.Decoder.Decode(conn, msg)
		//n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("TCP Error: %s\n", err)
		}
		msg.From = conn.RemoteAddr()
		fmt.Printf("Message: %v\n", msg)
		// if err := t.Decoder.Decode(conn, msg); err != nil {
		// 	fmt.Printf("TCP Error: %s", err);
		// 	continue;
		// } 
	}

	
}

