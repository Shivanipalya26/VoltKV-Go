package server

import (
	"bufio"
	"fmt"
	"go_redis/internals/cmd"
	"go_redis/internals/resp"
	"go_redis/internals/store"
	"log"
	"net"
)

type Server struct {
	address string
	store *store.Store
}

func NewServer(address string, s *store.Store) *Server {
	return &Server{address: address, store: s}
}

func (srv *Server) Start() error {
	ln, err := net.Listen("tcp", srv.address)
	if err != nil {
		return err
	}
	fmt.Println("Server listening on ", srv.address)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go srv.handleConnection(conn)
	}
}

func (srv *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	respReader := resp.NewResp(reader)

	for {
		value, err := respReader.ReadValue()
		log.Println(value)
		if err != nil {
			fmt.Fprintf(conn, "-ERR %v\r\n", err)
			return 
		}

		if value.Typ != "array" || len(value.Array) == 0 {
			fmt.Fprintf(conn, "-ERR invalid command format\r\n")
			continue
		}

		args := make([]string, 0, len(value.Array))
		valid := true
		for _, v := range value.Array {
			if v.Typ == "bulk" {
				args = append(args, v.Bulk)
			} else if v.Typ == "string" {
				args = append(args, v.Str)
			} else {
				fmt.Fprintf(conn, "-ERR unsupported argument type\r\n")
				valid = false
				break
			}
		}

		if !valid {
			continue
		}

		cmd.Execute(args, srv.store, conn)		
	}
}