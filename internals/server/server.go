package server

import (
	"bufio"
	"fmt"
	"go_redis/internals/store"
	"net"
	"strconv"
	"strings"
)

func Start(addr string, s *store.Store) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Println("Server listening on ", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn, s)
	}
}

func handleConnection(conn net.Conn, s *store.Store) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Fields(line)

		if len(tokens) == 0 {
			continue
		}

		cmd := strings.ToUpper(tokens[0])

		switch cmd {
		case "PING":
			if len(tokens) == 1 {
				fmt.Fprintln(conn, "PONG")
			} else {
				fmt.Fprintln(conn, tokens[1])
			}

		case "SET":
			if len(tokens) != 3 {
				fmt.Fprintln(conn, "ERR: SET requires 3 arguments")
				continue
			}
			s.Set(tokens[1], tokens[2])
			fmt.Fprintln(conn, "OK")
		
		case "GET":
			if len(tokens) != 2 {
				fmt.Fprintln(conn, "ERR: GET requires 2 arguments")
				continue
			}
			if val, ok := s.Get(tokens[1]); ok {
				fmt.Fprintln(conn, val)
			} else {
				fmt.Fprintln(conn, "null")
			}

		case "MSET":
			if len(tokens) % 2 != 1 {
				fmt.Fprintln(conn, "ERR: MSET requires even numbers of key-value pairs")
				continue
			}
			for i := 1; i < len(tokens); i += 2 {
				key := tokens[i]
				val := tokens[i+1]
				s.Set(key, val)
			}
			fmt.Fprintln(conn, "Ok")

		case "MGET":
			if len(tokens) < 2 {
				fmt.Fprintln(conn, "ERR: MGET requires at least one key")
				continue
			}
			results := []string{}
			for _, key := range tokens[1:] {
				if val, ok := s.Get(key); ok {
					results = append(results, val)
				} else {
					results = append(results, "null")
				}
			}
			fmt.Fprintln(conn, strings.Join(results, " "))

		case "HSET":
			if len(tokens) < 4 || len(tokens)%2 != 0 {
				fmt.Fprintln(conn, "ERR: HSET requires key followed by field value pairs")
				continue
			}
			key := tokens[1]
			fields := tokens[2:]
			
			fieldMap := make(map[string]string)
			for i := 0; i < len(fields); i += 2 {
				fieldMap[fields[i]] = fields[i+1]
			}
			s.HSet(key, fieldMap)
			fmt.Fprintln(conn, "OK")

		case "HGET":
			if len(tokens) != 3 {
				fmt.Fprintln(conn, "ERR: HGET requires 3 arguments")
				continue
			}
			if val, ok := s.HGet(tokens[1], tokens[2]); ok {
				fmt.Fprintln(conn, val)
			} else {
				fmt.Fprintln(conn, "null")
			}

		case "HGETALL":
			if len(tokens) != 2 {
				fmt.Fprintln(conn, "ERR: HGETALL requires 2 arguments")
				continue
			}
			all := s.HGetAll(tokens[1])
			if all == nil {
				fmt.Fprintln(conn, "null")
				continue
			}

			for k, v := range all {
				fmt.Fprintf(conn, "%s: %s\n", k, v)
			}

		case "DEL":
			if len(tokens) != 2 {
				fmt.Fprintln(conn, "ERR: DEL requires 2 arguments")
				continue
			}
			if deleted := s.Del(tokens[1]); deleted {
				fmt.Fprintln(conn, "(integer) 1")
			} else {
				fmt.Fprintln(conn, "(integer) 0")
			}

		case "EXISTS":
			if len(tokens) != 2 {
				fmt.Fprintln(conn, "ERR: EXISTS requires 2 arguments")
				continue
			}
			if s.Exists(tokens[1]) {
				fmt.Fprintln(conn, "(integer) 1")
			} else {
				fmt.Fprintln(conn, "(integer) 0")
			}

		case "EXPIRE":
			if len(tokens) != 3 {
				fmt.Fprintln(conn, "ERR: EXPIRE requires 3 arguments")
				continue
			}
			seconds, err := strconv.Atoi(tokens[2])
			if err != nil || seconds < 0 {
				fmt.Fprintln(conn, "ERR: invalid seconds")
				continue
			}
			if ok := s.Expire(tokens[1], seconds); ok {
				fmt.Fprintln(conn, "(integer) 1")
			} else {
				fmt.Fprintln(conn, "(integer) 0")
			}

		default : 
			fmt.Fprintln(conn, "ERR: unknown cmd", cmd)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(conn, "ERR: reading from connection : ", err)
	}
}