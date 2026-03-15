package p2p

import (
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
	var err error

	defer func(){
		fmt.Printf("Dropping peer connection: %s", err)
		conn.Close()
	}()
	
	peer := NewTCPPeer(conn, true)
	//fmt.Printf("new connection at %v\n", peer)

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

