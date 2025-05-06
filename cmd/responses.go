package cmd

import (
	"fmt"
	"net"
)

func writeOk(conn net.Conn) {
	fmt.Fprintf(conn, "+Ok\r\n")
}

func writeString(conn net.Conn, s string) {
	fmt.Fprintf(conn, "+%s\r\n", s)
}

func writeBulkString(conn net.Conn, s string) {
	fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(s), s)
}

func writeNullBulkString(conn net.Conn) {
	fmt.Fprintf(conn, "$-1\r\n")
}

func writeInteger(conn net.Conn, n int) {
	fmt.Fprintf(conn, ":%d\r\n", n)
}

func writeError(conn net.Conn, errMsg string) {
	fmt.Fprintf(conn, "-ERR %s\r\n", errMsg)
}

func writeArray(conn net.Conn, data []string) {
	fmt.Fprintf(conn, "*%d\r\n", len(data))
	for _, val := range data {
		fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(val), val)
	}
}

func writeNullArray(conn net.Conn) {
	fmt.Fprint(conn, "*-1\r\n")
}
