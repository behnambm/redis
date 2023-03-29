package main

import (
	"bufio"
	"fmt"
	"strconv"
)

type Type byte

var (
	SimpleStringType Type = '+'
	BulkStringType   Type = '$'
	ArrayType        Type = '*'
)

type Value struct {
	typ   Type
	bytes []byte
	array []Value
}

func (v Value) Array() []Value {
	if v.typ == ArrayType {
		return v.array
	}

	return []Value{}
}

func (v Value) String() string {
	if v.typ == SimpleStringType || v.typ == BulkStringType {
		return string(v.bytes)
	}

	return ""
}
func DecodeRESP(byteStream *bufio.Reader) (Value, error) {
	firstByte, err := byteStream.ReadByte()
	if err != nil {
		return Value{}, err
	}
	switch string(firstByte) {
	case "+":
		return decodeSimpleString(byteStream)
	case "$":
		return decodeBulkString(byteStream)
	case "*":
		return decodeArray(byteStream)
	default:
		return Value{}, fmt.Errorf("invalid first byte(data type)")
	}
}

func decodeBulkString(byteStream *bufio.Reader) (Value, error) {
	b, err := readUntilCRLF(byteStream)
	if err != nil {
		return Value{}, fmt.Errorf("bulk string - invalid first section: %w", err)
	}
	stringLength, err := strconv.Atoi(string(b))
	if err != nil {
		return Value{}, fmt.Errorf("bulk string - invalid string length: %w", err)
	}
	bulkString, err := readUntilCRLF(byteStream)
	if err != nil {
		return Value{}, fmt.Errorf("bulk string - invalid second section: %w", err)
	}
	if len(bulkString) != stringLength {
		return Value{}, fmt.Errorf("bulk string - string and length doesn't match")
	}

	return Value{
		typ:   BulkStringType,
		bytes: bulkString,
	}, nil
}

func decodeSimpleString(byteStream *bufio.Reader) (Value, error) {
	b, err := readUntilCRLF(byteStream)
	if err != nil {
		return Value{}, fmt.Errorf("invalid byte stream in simple string: %w", err)
	}

	return Value{
		typ:   SimpleStringType,
		bytes: b,
	}, nil
}

func decodeArray(byteStream *bufio.Reader) (Value, error) {
	arrayCountB, err := readUntilCRLF(byteStream)
	if err != nil {
		return Value{}, fmt.Errorf("array - invalid array length: %w", err)
	}
	arrayCount, err := strconv.Atoi(string(arrayCountB))
	if err != nil {
		return Value{}, fmt.Errorf("array - can't parse array length to int: %w", err)
	}

	valueArray := []Value{}
	for i := 0; i < arrayCount; i++ {
		tmpValue, err := DecodeRESP(byteStream)
		if err != nil {
			fmt.Println("array - element number ", i, " - %w", err)
		}
		valueArray = append(valueArray, tmpValue)
	}

	return Value{
		typ:   ArrayType,
		array: valueArray,
	}, nil
}

func readUntilCRLF(bytes *bufio.Reader) ([]byte, error) {
	buf := []byte{}
	for {
		b, err := bytes.ReadBytes('\n')
		if err != nil {
			return []byte{}, err
		}
		buf = append(buf, b...)
		if len(buf) >= 2 && buf[len(buf)-2] == '\r' {
			break
		}
	}

	return buf[:len(buf)-2], nil
}
