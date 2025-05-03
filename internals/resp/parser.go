package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING = '+'
	ARRAY = '*'
	INTEGER = ':'
	BULK = '$'
	ERROR = '-'
)

var (
	ErrUnexpectedType = errors.New("unexpected RESP type")
	ErrInvalidSyntax = errors.New("unexpected RESP syntax")
)

type Resp struct {
	r *bufio.Reader
}

func NewResp(r *bufio.Reader) *Resp {
	return &Resp{r : r}
}

type Value struct {
	Typ string
	Str string
	Array []Value
	Num int 
	Bulk string
	Err string
}

func (resp *Resp) ReadValue() (Value, error) {
	dataType, err := resp.r.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch dataType {
	case ARRAY: 
		return resp.readArray()

	case BULK: 
		return resp.readBulk()

	case STRING: 
		return resp.readString()

	case INTEGER:
		return resp.readNum()
	
	case ERROR: 
		return resp.readError()
	
	default:
		return Value{}, ErrUnexpectedType
	}	
}

func (resp *Resp) readLine() ([]byte, int, error) {
	var line []byte
	var n int

	for {
		b, err := resp.r.ReadByte()
		if err != nil {
			return nil, n, err
		}
		line = append(line, b)
		n++
		if len(line) >= 2 && line[len(line)-2] == '\r' && line[len(line)-1] == '\n' {
			break
		}
	}
	return line[:len(line)-2], n, nil
}

func (resp *Resp) readInt() (int, int, error) {
	str, n, err := resp.readLine()
	if err != nil {
		return 0, 0, err
	}
	num, err := strconv.ParseInt(string(str), 10, 64)
	if err != nil {
		return 0, n, fmt.Errorf("%w: %v", ErrInvalidSyntax, err)
	}
	return int(num), n, nil
}

func (resp *Resp) readArray() (Value, error) {
	v := Value{Typ: "array"}
	n, _, err := resp.readInt()
	if err != nil {
		return v, err
	}

	if n < 0 {
		return Value{Typ: "null"}, nil
	}

	v.Array = make([]Value, n)
	for i := 0; i < n; i++ {
		val, err := resp.ReadValue()
		if err != nil {
			return v, err
		} 
		v.Array[i] = val
	}
	return v, nil
}

func (resp *Resp) readBulk() (Value, error) {
	v := Value{Typ: "bulk"}
	len, _, err := resp.readInt()
	if err != nil {
		return v, err
	}
	if len < 0 {
		return Value{Typ: "null"}, nil
	}
	buf := make([]byte, len+2)
	if _, err := io.ReadFull(resp.r, buf); err != nil {
		return v, err
	}
	if buf[len] != '\r' || buf[len+1] != '\n' {
		return v, ErrInvalidSyntax
	}
	v.Str = string(buf[:len])
	v.Bulk = v.Str
	// resp.r.ReadLine()
	return v, nil
}

func (resp *Resp) readString() (Value, error) {
	v := Value{Typ: "string"}
	str, _, err := resp.readLine()
	if err != nil {
		return v, err
	}
	v.Str = string(str)
	return v, nil
}

func (resp *Resp) readNum() (Value, error) {
	num, _, err := resp.readInt()
	if err != nil {
		return Value{}, err
	}
	return Value{Typ: "integer", Num: num}, nil
}

func (resp *Resp) readError() (Value, error) {
	v := Value{Typ: "error"}
	str, _, err := resp.readLine()
	if err != nil {
		return v, err
	}
	v.Err = string(str)
	return v, nil
}
