package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	myCache := New()

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		log.Println("ESTABLISHED A NEW CONNECTION: ", conn.RemoteAddr())

		go handleConnection(conn, myCache)
	}
}

func handleConnection(conn net.Conn, cache *MemoryCache) {
	defer conn.Close()
	for {
		value, err := DecodeRESP(bufio.NewReader(conn))
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Println("decode err: ", err)
			return
		}
		command := value.Array()[0].String()

		switch strings.ToLower(command) {
		case "ping":
			conn.Write([]byte("+PONG\r\n"))
		case "echo":
			args := value.Array()[1].String()
			WriteBulkString(conn, args)
		case "get":
			key := value.Array()[1].String()
			dataValue, cacheErr := cache.Get(key)
			if cacheErr != nil {
				if cacheErr == ValueExpiredError {
					cache.Delete(key)
				}
				conn.Write([]byte("$-1\r\n"))
			} else {
				dataValueStr := dataValue.value
				WriteBulkString(conn, dataValueStr)
			}
		case "set":
			key := value.Array()[1].String()
			if len(value.Array()) < 3 {
				WriteError(conn, "not enough arguments for SET command")
				continue
			}
			valueToStore := value.Array()[2].String()
			if len(value.Array()) == 5 {
				expiryArg := value.Array()[3].String()
				if strings.ToLower(expiryArg) != "px" {
					WriteError(conn, fmt.Sprintf("invalid SET argument: %s", expiryArg))
				}
				expirationTime := value.Array()[4].String()
				expirationTimeInt, convErr := strconv.Atoi(expirationTime)
				if convErr != nil {
					WriteError(conn, fmt.Sprintf("invalid expiration time: %s", expirationTime))
				}
				cacheValueToStoreWithExpiry := CacheValue{
					value:      valueToStore,
					expiration: time.Now().Add(time.Millisecond * time.Duration(expirationTimeInt)),
				}
				cache.Set(key, cacheValueToStoreWithExpiry)
			} else {
				cacheValueToStore := CacheValue{
					value: valueToStore,
				}
				cache.Set(key, cacheValueToStore)
			}

			conn.Write([]byte("+OK\r\n"))
		default:
			WriteError(conn, fmt.Sprintf("invalid command: %s", command))
		}
	}
}

func WriteError(w io.Writer, msg string) {
	w.Write([]byte(fmt.Sprintf("-ERR %s\r\n", msg)))
}

func WriteBulkString(w io.Writer, msg string) {
	w.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(msg), msg)))
}
