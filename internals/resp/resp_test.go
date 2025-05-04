package resp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestRespReadValue(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected Value
		wantErr  error
	}{
		{
			name:     "Simple String",
			input:    []byte("+OK\r\n"),
			expected: Value{Typ: "string", Str: "OK"},
			wantErr:  nil,
		},
		{
			name:     "Empty String",
			input:    []byte("+\r\n"),
			expected: Value{Typ: "string", Str: ""},
			wantErr:  nil,
		},
		{
			name:     "Bulk String",
			input:    []byte("$5\r\nhello\r\n"),
			expected: Value{Typ: "bulk", Bulk: "hello"},
			wantErr:  nil,
		},
		{
			name:     "Empty Bulk String",
			input:    []byte("$0\r\n\r\n"),
			expected: Value{Typ: "bulk", Bulk: ""},
			wantErr:  nil,
		},
		{
			name:     "Null Bulk String",
			input:    []byte("$-1\r\n"),
			expected: Value{Typ: "null"},
			wantErr:  nil,
		},
		{
			name:     "Integer",
			input:    []byte(":1000\r\n"),
			expected: Value{Typ: "integer", Num: 1000},
			wantErr:  nil,
		},
		{
			name:     "Negative Integer",
			input:    []byte(":-123\r\n"),
			expected: Value{Typ: "integer", Num: -123},
			wantErr:  nil,
		},
		{
			name:     "Simple Error",
			input:    []byte("-Error something\r\n"),
			expected: Value{Typ: "error", Err: "Error something"},
			wantErr:  nil,
		},
		{
			name:  "Array of Bulk Strings",
			input: []byte("*2\r\n$5\r\nhello\r\n$6\r\ngolang\r\n"),
			expected: Value{Typ: "array",
				Array: []Value{
					{Typ: "bulk", Bulk: "hello"},
					{Typ: "bulk", Bulk: "golang"},
				},
			},
			wantErr: nil,
		},
		{
			name: "Empty Array",
			input: []byte("*0\r\n"),
			expected: Value{
				Typ: "array",
				Array: []Value{},
			},
			wantErr: nil,
		},
		{
			name: "Null Array",
			input: []byte("*-1\r\n"),
			expected: Value{Typ: "null"},
			wantErr: nil,
		},
		{
			name: "Nested Array",
			input: []byte("*2\r\n*2\r\n:1\r\n:2\r\n*2\r\n+a\r\n-b\r\n"),
			expected: Value{
				Typ: "array",
				Array: []Value{
					{
						Typ: "array",
						Array: []Value{
							{Typ: "integer", Num: 1},
							{Typ: "integer", Num: 2},
						},
					}, {
						Typ: "array",
						Array: []Value{
							{Typ: "string", Str: "a"},
							{Typ: "error", Err: "b"},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Array with Null Bulk String",
			input: []byte("*3\r\n$3\r\nfoo\r\n$-1\r\n$3\r\nbar\r\n"),
			expected: Value{
				Typ: "array",
				Array: []Value{
					{Typ: "bulk", Bulk: "foo"},
					{Typ: "null"},
					{Typ: "bulk", Bulk: "bar"},
				},
			},
			wantErr: nil,
		},
		{
			name: "Array with Mixed Types",
			input: []byte("*3\r\n+hello\r\n:42\r\n$6\r\ngolang\r\n"),
			expected: Value{
				Typ: "array",
				Array: []Value{
					{Typ: "string", Str: "hello"},
					{Typ: "integer", Num: 42},
					{Typ: "bulk", Bulk: "golang"},
				},
			},
			wantErr: nil,
		},
		{
			name: "Invalid Type",
			input: []byte("?invalid\r\n"),
			expected: Value{},
			wantErr: ErrUnexpectedType,
		},
		{
			name: "Malformed Type",
			input: []byte(":notanumber\r\n"),
			expected: Value{},
			wantErr: ErrInvalidSyntax,
		},
		{
			name: "Incomplete Array",
			input: []byte("*2\r\n$5\r\nhello\r\n"),
			expected: Value{},
			wantErr: io.EOF,
		},
		{
			name: "Missing CRLF in BULK String",
			input: []byte("$5\r\nhello"),
			expected: Value{},
			wantErr: fmt.Errorf("unexpected RESP syntax: expected CRLF, got EOF"),
		},
		{
			name: "Large Bulk String",
			input: func() []byte {
				data := bytes.Repeat([]byte("a"), 10000)
				buf := bytes.NewBuffer(nil)
				buf.WriteString("$10000\r\n")
				buf.Write(data)
				buf.WriteString("\r\n")
				return buf.Bytes()
			}(),
			expected: Value{
				Typ: "bulk",
				Bulk: string(bytes.Repeat([]byte("a"), 10000)),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bufio.NewReader(bytes.NewReader(tt.input))
			resp := NewResp(r)
			val, err := resp.ReadValue()

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr.Error())
				}
			
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %q, got %q", tt.wantErr.Error(), err.Error())
				}
				return
			}	
			
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(val, tt.expected) {
				t.Errorf("expected: %+v, got: %+v", tt.expected, val)
			}
		})
	}
}
