package cmd

import (
	"fmt"
	"go_redis/internals/store"
	"net"
	"strconv"
	"strings"
	"time"
)

func Execute(args []string, s *store.Store, conn net.Conn) {
	fmt.Printf("ARGS: %#v\n", args)

	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		writeError(conn, "missing command")
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

	case "LPUSH":
		handleLPush(args, s, conn)

	case "RPUSH":
		handleRPush(args, s, conn)

	case "LPOP":
		handleLPop(args, s, conn)

	case "RPOP":
		handleRPop(args, s, conn)

	case "BLPOP":
		handleBLPop(args, s, conn)

	default:
		fmt.Fprintf(conn, "-ERR unknown command '%s'\r\n", cmd)
	}
}

func handlePing(args []string, conn net.Conn) {
	if len(args) == 1 {
		writeString(conn, "PONG")
	} else {
		writeBulkString(conn, args[1])
	}
}

func handleSet(args []string, s *store.Store, conn net.Conn) {
	if len(args) != 3 {
		writeError(conn, "wrong no. of arguments for 'set'")
		return
	}
	s.Set(args[1], args[2])
	writeOk(conn)
}

func handleGet(args []string, s *store.Store, conn net.Conn) {
	if len(args) != 2 {
		writeError(conn, "wrong no. of arguments for 'get'")
		return
	}
	if val, ok := s.Get(args[1]); ok {
		writeBulkString(conn, val)
	} else {
		writeNullBulkString(conn)
	}
}

func handleMSet(args []string, s *store.Store, conn net.Conn) {
	if len(args)%2 != 1 {
		writeError(conn, "wrong no. of arguments for 'mset'")
		return
	}
	for i := 1; i < len(args); i += 2 {
		key := args[i]
		val := args[i+1]
		s.Set(key, val)
	}
	writeOk(conn)
}

func handleMGet(args []string, s *store.Store, conn net.Conn) {
	if len(args) < 2 {
		writeError(conn, "wrong no. of arguments for 'mget'")
		return
	}
	fmt.Fprintf(conn, "*%d\r\n", len(args)-1)
	for _, key := range args[1:] {
		if val, ok := s.Get(key); ok {
			writeBulkString(conn, val)
		} else {
			writeNullBulkString(conn)
		}
	}
}

func handleHSet(args []string, s *store.Store, conn net.Conn) {
	if len(args) < 4 || len(args)%2 != 0 {
		writeError(conn, "wrong no. of arguments for 'hset'")
		return
	}
	key := args[1]
	fields := args[2:]

	fieldMap := make(map[string]string)
	for i := 0; i < len(fields); i += 2 {
		fieldMap[fields[i]] = fields[i+1]
	}
	s.HSet(key, fieldMap)
	writeOk(conn)
}

func handleHGet(args []string, s *store.Store, conn net.Conn) {
	if len(args) != 3 {
		writeError(conn, "wrong no. of arguments for 'hget'")
		return
	}
	if val, ok := s.HGet(args[1], args[2]); ok {
		writeBulkString(conn, val)
	} else {
		writeNullBulkString(conn)
	}
}

func handleHGetAll(args []string, s *store.Store, conn net.Conn) {
	if len(args) != 2 {
		writeError(conn, "wrong no. of arguments for 'hgetall'")
		return
	}
	all := s.HGetAll(args[1])
	if all == nil {
		writeNullBulkString(conn)
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
		writeError(conn, "wrong no. of arguments for 'del'")
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
		writeError(conn, "wrong no. of arguments for 'exists'")
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
		writeError(conn, "wrong no. of arguments for 'expire'")
		return
	}
	seconds, err := strconv.Atoi(args[2])
	if err != nil || seconds < 0 {
		writeError(conn, "invalid seconds")
		return
	}
	if ok := s.Expire(args[1], seconds); ok {
		fmt.Fprint(conn, ":1\r\n")
	} else {
		fmt.Fprint(conn, ":0\r\n")
	}
}

func handleLPush(args []string, s *store.Store, conn net.Conn) {
	if len(args) < 3 {
		writeError(conn, "wrong no. of arguments for 'lpush'")
		return
	}
	key := args[1]
	values := args[2:]

	count := s.LPush(key, values...)
	writeInteger(conn, count)
}

func handleRPush(args []string, s *store.Store, conn net.Conn) {
	if len(args) < 3 {
		writeError(conn, "wrong no. of arguments for 'rpush'")
		return
	}

	key := args[1]
	values := args[2:]

	count := s.RPush(key, values...)
	writeInteger(conn, count)
}

func handleLPop(args []string, s *store.Store, conn net.Conn) {
	if len(args) != 2 {
		writeError(conn, "wrong no. of arguments for 'lpop'")
		return
	}

	key := args[1]
	val, ok := s.LPop(key)
	if !ok {
		writeNullBulkString(conn)
		return
	}
	writeBulkString(conn, val)
}

func handleRPop(args []string, s *store.Store, conn net.Conn) {
	if len(args) != 2 {
		writeError(conn, "wrong no. of arguments for 'rpop'")
		return
	}

	key := args[1]
	val, ok := s.RPop(key)
	if !ok {
		writeNullBulkString(conn)
		return
	}
	writeBulkString(conn, val)
}

func handleBLPop(args []string, s *store.Store, conn net.Conn) {
	if len(args) < 2 {
		writeError(conn, "wrong no. of arguments for 'blpop'")
		return
	}

	keys := args[:len(args)-1]
	timeout, err := strconv.Atoi(args[len(args)-1])

	if err != nil {
		writeError(conn, "timeout is not an integer")
		return
	}

	waitCh := make(chan [2]string, 1)

	for _, key := range keys {
		val, ok := s.LPop(key)
		if ok {
			writeArray(conn, []string{key, val})
			return
		}
	}

	for _, key := range keys {
		s.RegisterWaiter(key, waitCh)
	}

	select {
	case result := <-waitCh:
		writeArray(conn, []string{result[0], result[1]})

	case <-time.After(time.Duration(timeout) * time.Second):
		writeNullArray(conn)
	}
}
