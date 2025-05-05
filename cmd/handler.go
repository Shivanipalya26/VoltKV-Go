package cmd

import (
	"fmt"
	"go_redis/internals/store"
	"net"
	"strconv"
	"strings"
)

func Execute(args []string, s *store.Store, conn net.Conn) {
	fmt.Printf("ARGS: %#v\n", args)

	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		fmt.Fprint(conn, "-ERR empty command\r\n")
		return
	}

	cmd := strings.ToUpper(args[0])

	switch cmd {
	case "PING":
		handlePing(args, conn)

	case "SET":
		handleSet(args, s, conn)

	case "GET":
		handleGet(args, s, conn)

	case "MSET":
		handleMSet(args, s, conn)

	case "MGET":
		handleMGet(args, s, conn)

	case "HSET":
		handleHSet(args, s, conn)

	case "HGET":
		handleHGet(args, s, conn)

	case "HGETALL":
		handleHGetAll(args, s, conn)

	case "DEL":
		handleDel(args, s, conn)

	case "EXISTS":
		handleExists(args, s, conn)

	case "EXPIRE":
		handleExpire(args, s, conn)

	default:
		fmt.Fprintf(conn, "-ERR unknown command '%s'\r\n", cmd)
	}
}

func handlePing(args []string, conn net.Conn) {
	if len(args) == 1 {
		fmt.Fprint(conn, "+PONG\r\n")
	} else {
		msg := args[1]
		fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(msg), msg)
	}
}

func handleSet(args []string, s *store.Store, conn net.Conn) {
	if len(args) != 3 {
		fmt.Fprint(conn, "-ERR: SET requires 3 arguments\r\n")
		return
	}
	s.Set(args[1], args[2])
	fmt.Fprint(conn, "+OK\r\n")
}

func handleGet(args []string, s *store.Store, conn net.Conn) {
	if len(args) != 2 {
		fmt.Fprint(conn, "-ERR: GET requires 2 arguments\r\n")
		return
	}
	if val, ok := s.Get(args[1]); ok {
		fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(val), val)
	} else {
		fmt.Fprint(conn, "$-1\r\n")
	}
}

func handleMSet(args []string, s *store.Store, conn net.Conn) {
	if len(args)%2 != 1 {
		fmt.Fprint(conn, "-ERR: MSET requires even numbers of key-value pairs\r\n")
		return
	}
	for i := 1; i < len(args); i += 2 {
		key := args[i]
		val := args[i+1]
		s.Set(key, val)
	}
	fmt.Fprint(conn, "+Ok\r\n")
}

func handleMGet(args []string, s *store.Store, conn net.Conn) {
	if len(args) < 2 {
		fmt.Fprint(conn, "-ERR: MGET requires at least one key\r\n")
		return
	}
	fmt.Fprintf(conn, "*%d\r\n", len(args)-1)
	for _, key := range args[1:] {
		if val, ok := s.Get(key); ok {
			fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(val), val)
		} else {
			fmt.Fprint(conn, "$-1\r\n")
		}
	}
}

func handleHSet(args []string, s *store.Store, conn net.Conn) {
	if len(args) < 4 || len(args)%2 != 0 {
		fmt.Fprint(conn, "-ERR: HSET requires key followed by field value pairs\r\n")
		return
	}
	key := args[1]
	fields := args[2:]

	fieldMap := make(map[string]string)
	for i := 0; i < len(fields); i += 2 {
		fieldMap[fields[i]] = fields[i+1]
	}
	s.HSet(key, fieldMap)
	fmt.Fprint(conn, ":1\r\n")
}

func handleHGet(args []string, s *store.Store, conn net.Conn) {
	if len(args) != 3 {
		fmt.Fprint(conn, "-ERR: HGET requires 3 arguments\r\n")
		return
	}
	if val, ok := s.HGet(args[1], args[2]); ok {
		fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(val), val)
	} else {
		fmt.Fprint(conn, "$-1\r\n")
	}
}

func handleHGetAll(args []string, s *store.Store, conn net.Conn) {
	if len(args) != 2 {
		fmt.Fprint(conn, "-ERR: HGETALL requires 2 arguments\r\n")
		return
	}
	all := s.HGetAll(args[1])
	if all == nil {
		fmt.Fprint(conn, "$-1\r\n")
		return
	}

	fmt.Fprintf(conn, "*%d\r\n", len(all)*2)
	for k, v := range all {
		fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(k), k)
		fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(v), v)

	}
}

func handleDel(args []string, s *store.Store, conn net.Conn) {
	if len(args) != 2 {
		fmt.Fprint(conn, "-ERR: DEL requires 2 arguments\r\n")
		return
	}
	if deleted := s.Del(args[1]); deleted {
		fmt.Fprint(conn, ":1\r\n")
	} else {
		fmt.Fprint(conn, ":0\r\n")
	}
}

func handleExists(args []string, s *store.Store, conn net.Conn) {
	if len(args) != 2 {
		fmt.Fprint(conn, "-ERR: EXISTS requires 2 arguments\r\n")
		return
	}
	if s.Exists(args[1]) {
		fmt.Fprint(conn, ":1\r\n")
	} else {
		fmt.Fprint(conn, ":0\r\n")
	}
}

func handleExpire(args []string, s *store.Store, conn net.Conn) {
	if len(args) != 3 {
		fmt.Fprint(conn, "-ERR: EXPIRE requires 3 arguments\r\n")
		return
	}
	seconds, err := strconv.Atoi(args[2])
	if err != nil || seconds < 0 {
		fmt.Fprint(conn, "-ERR: invalid seconds\r\n")
		return
	}
	if ok := s.Expire(args[1], seconds); ok {
		fmt.Fprint(conn, ":1\r\n")
	} else {
		fmt.Fprint(conn, ":0\r\n")
	}
}
