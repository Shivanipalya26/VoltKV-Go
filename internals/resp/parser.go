package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
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

	log.Printf("Data type byte: %c", dataType)

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
	length, _, err := resp.readInt()
	if err != nil {
		return v, err
	}

	// log.Printf("Array length: %d", length)

	if length < 0 {
		return Value{Typ: "null"}, nil
	}

	v.Array = make([]Value, length)

	for i := 0; i < length; i++ {
		val, err := resp.ReadValue()
		if err != nil {
			// log.Printf("Failed reading element %d: %v", i, err)
			return v, err
		} 
		// log.Printf("Array element %d: %+v", i, val) 
		v.Array[i] = val
	}
	return v, nil
}

func (resp *Resp) readBulk() (Value, error) {
	v := Value{Typ: "bulk"}
	length, _, err := resp.readInt()
	if err != nil {
		return v, err
	}
	if length < 0 {
		return Value{Typ: "null"}, nil
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(resp.r, data); err != nil {
		return v, fmt.Errorf("%w: failed to read bulk data: %v", ErrInvalidSyntax, err)
	}

	crlf := make([]byte, 2)
	if _, err := io.ReadFull(resp.r, crlf); err != nil {
		if err == io.EOF {
        return v, fmt.Errorf("unexpected RESP syntax: expected CRLF, got EOF")
    }
		return v, fmt.Errorf("%w: failed to read CRLF after bulk data: %v", ErrInvalidSyntax, err)
	}
	if crlf[0] != '\r' || crlf[1] != '\n' {
		return v, fmt.Errorf("%w: invalid line ending in bulk string", ErrInvalidSyntax)
	}

	v.Bulk = string(data)

	// log.Printf("Read bulk string of length %d: %q", length, v.Bulk)
	// log.Printf("CRLF after bulk: %q", crlf)

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
