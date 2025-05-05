package server

import (
	"bufio"
	"fmt"
	"go_redis/cmd"
	"go_redis/internals/resp"
	"go_redis/internals/store"
	"log"
	"net"
)

type Peer struct {
	conn    net.Conn
	cmdChan chan Command
	reader  *resp.Resp
	writer  *bufio.Writer
	name    string
}

func NewPeer(conn net.Conn, cmdChan chan Command) *Peer {
	return &Peer{
		conn:    conn,
		cmdChan: cmdChan,
		reader:  resp.NewResp(bufio.NewReader(conn)),
		writer:  bufio.NewWriter(conn),
		name:    conn.RemoteAddr().String(),
	}
}

func (p *Peer) ReadLoop(srv *Server) {
	defer func() {
		p.conn.Close()
		srv.removePeerChan <- p
	}()

	for {
		val, err := p.reader.ReadValue()
		if err != nil {
			p.WriteError("ERR " + err.Error())
			return
		}
		cmd := Command{
			Peer: p,
			Args: val,
		}
		srv.cmdChan <- cmd
	}
}

func (p *Peer) Handle(args []string, s *store.Store) {
	log.Printf("[Peer %s] Executing command: %v", p.name, args)
	cmd.Execute(args, s, p.conn)
}

func (p *Peer) WriteError(message string) {
	fmt.Fprintf(p.conn, "-ERR %s\r\n", message)
}
