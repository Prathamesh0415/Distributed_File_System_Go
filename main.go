package main

import (
	//"fmt"
	"log"

	"github.com/Prathamesh0415/fileserver/p2p"
) 

func main(){
	tr := p2p.NewTCPTransport(":4000")
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal((err))
	}

	select {}
}

