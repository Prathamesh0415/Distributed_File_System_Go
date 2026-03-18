package p2p

import (
	"errors"
	"fmt"
	"net"
	//"sync"
)

type TCPPeer struct{
	conn net.Conn
	outbound bool
}

type TCPTransportOpts struct {
	ListenAddress string
	HandshakeFunc HandshakeFunc
	Decoder Decoder
	OnPeer func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch chan RPC
}

func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)

	return nil
} 

func (r *TCPTransport) Consume() <- chan RPC{
	return r.rpcch 
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
		rpcch: make(chan RPC),
	}
}

func (t *TCPTransport) ListenAndAccept() error{
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		return err
	} 

	fmt.Printf("TCP Server listening on port: %s\n", t.ListenAddress)

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) Close() error {
	return t.listener.Close()

}

func (t *TCPTransport) startAcceptLoop() {
	for{
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return 
		}
		if err != nil {
			fmt.Printf("Error at start accept loop function: %s\n", err)
		}

		fmt.Printf("new connection at %v\n", conn)

		go t.handleConn(conn, true)
	}
}



func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error

	defer func(){
		fmt.Printf("Dropping peer connection: %s", err)
		conn.Close()
	}()
	
	peer := NewTCPPeer(conn, outbound)
	

	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	//read loop
	rpc := RPC{};
	for {
		err := t.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}
		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
	
	}
}

