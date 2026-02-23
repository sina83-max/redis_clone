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
		// We pass nil for Aof because we don't want to
		// write to the AOF file while we are reading FROM it
		executeCommand(value, db, nil)
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
	resp := NewResp(conn)
	writer := NewWriter(conn)
	for {
		// Read
		value, err := resp.Read()
		if err != nil {
			return // Client disconnected
		}

		// Execute
		result := executeCommand(value, db, aof)

		// Write
		writer.Write(result)
	}
}

// executeCommand acts like the "controller" of the app
func executeCommand(value Value, db *DB, aof *Aof) Value {
	// 1. Basic Validation
	if value.typ != "array" || len(value.array) == 0 {
		fmt.Println("Invalid request, expected non-empty array")
		return Value{
			typ: "err",
			str: "ERR unknown command format",
		}
	}

	// 3. Extract command and Args
	command := strings.ToUpper(value.array[0].bulk)
	args := value.array[1:]

	// 4. Lookup handler
	handler, ok := Handlers[command]
	if !ok {
		return Value{
			typ: "error",
			str: "ERR unknown command '" + command + "'",
		}
	}

	// 4. Persistance (AOF) - Only for write commands
	// Note: This list must be dynamic, not hardcoded
	if (command == "SET" || command == "HSET") && aof != nil {
		aof.Write(value)
	}

	// 5. Execute handler and send response
	return handler(args, db)
}
