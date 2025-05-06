package server

import (
	"fmt"
	"go_redis/internals/resp"
	"go_redis/internals/store"
	"log"
	"net"
	"sync"
)

type Server struct {
	address        string
	store          *store.Store
	cmdChan        chan Command
	peers          map[*Peer]bool
	addPeerChan    chan *Peer
	removePeerChan chan *Peer
	mu             sync.Mutex
}

type Command struct {
	Peer *Peer
	Args resp.Value
}

func NewServer(address string, s *store.Store) *Server {
	return &Server{
		address:        address,
		store:          s,
		cmdChan:        make(chan Command, 100),
		peers:          make(map[*Peer]bool),
		addPeerChan:    make(chan *Peer),
		removePeerChan: make(chan *Peer),
	}
}

func (srv *Server) Start() error {
	ln, err := net.Listen("tcp", srv.address)
	if err != nil {
		return err
	}
	fmt.Println("Server listening on ", srv.address)

	go srv.eventLoop()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		log.Printf("New connection from %s", conn.RemoteAddr())
		p := NewPeer(conn, srv.cmdChan)
		srv.addPeerChan <- p
		log.Printf("Peers %v", srv.peers)
		go p.ReadLoop(srv)
	}
}

func (srv *Server) eventLoop() {
	for {
		select {
		case p := <-srv.addPeerChan:
			srv.mu.Lock()
			srv.peers[p] = true
			srv.mu.Unlock()
			log.Printf("Added peer: %s", p.name)

		case p := <-srv.removePeerChan:
			srv.mu.Lock()
			delete(srv.peers, p)
			srv.mu.Unlock()
			log.Printf("Removed peer: %s", p.name)

		case cmd := <-srv.cmdChan:
			srv.handleConnection(cmd)
		}
	}
}

func (srv *Server) handleConnection(cmd Command) {
	args := parseArgs(cmd.Args)
	if len(args) == 0 {
		cmd.Peer.WriteError("Err invalid command")
		return
	}
	cmd.Peer.Handle(args, srv.store)
}

func parseArgs(value resp.Value) []string {
	args := make([]string, 0, len(value.Array))
	for _, v := range value.Array {
		switch v.Typ {
		case "bulk":
			args = append(args, v.Bulk)
		case "string":
			args = append(args, v.Str)
		}
	}
	return args
}
