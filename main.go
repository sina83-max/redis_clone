package main

import (
	"fmt"
	"net"
	"strings"
)

func main() {
	fmt.Println("listening on port :6379")

	// Initialize storage engine
	db := NewDB()

	// Create a new server
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	// Setup AOF (Persistance)
	aof, err := NewAof("database.aof")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer aof.Close()

	// AOF Recovery
	aof.Read(func(value Value) {
		if len(value.array) == 0 {
			return
		}
		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			return
		}

		handler(args, db)
	})

	// Main accept loop
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
			// Don't kill the server on a failed connection
		}

		// Handle connection in a seperate Goroutine
		// To handle Cuncurrency
		go handleClient(conn, db, aof)
	}
}

// handleClient manages the lifecycle of a single TCP connection
func handleClient(conn net.Conn, db *DB, aof *Aof) {
	defer conn.Close()

	for {
		resp := NewResp(conn)
		value, err := resp.Read()
		if err != nil {
			return
		}

		if value.typ != "array" || len(value.array) == 0 {
			fmt.Println("Invalid request, expected non-empty array")
			continue
		}

		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		writer := NewWriter(conn)

		handler, ok := Handlers[command]
		if !ok {
			writer.Write(Value{typ: "error", str: "ERR unknown command '" + command + "'"})
			continue
		}

		// Persist write commands to AOF
		if command == "SET" || command == "HSET" {
			aof.Write(value)
		}

		// Execute handler and send response
		result := handler(args, db)
		writer.Write(result)
	}
}
