package main

import "sync"

// a map whose keys are strings, and whose values are functions.
var Handlers = map[string]func([]Value) Value{
	"PING":    ping,
	"SET":     set,
	"GET":     get,
	"HSET":    hset,
	"HGET":    hget,
	"HGETALL": hgetall,
}

func ping(args []Value) Value {
	if len(args) == 0 {
		return Value{typ: "string", str: "Pong"}
	}

	return Value{typ: "string", str: args[0].bulk}
}

var SETs = map[string]string{}

// Many readers at the same time, Only one writer at a time
// makes sure only one goroutine (or thread) can modify shared data at a time.
var SETsMu = sync.RWMutex{}

func set(args []Value) Value {
	if len(args) != 2 {
		return Value{
			typ: "error",
			str: "ERR wrong number of arguments for 'SET' command",
		}
	}

	key := args[0].bulk
	value := args[1].bulk

	SETsMu.Lock()
	SETs[key] = value
	SETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func get(args []Value) Value {
	if len(args) != 1 {
		return Value{
			typ: "error",
			str: "ERR wrong number of arguments for 'GET' command",
		}
	}

	key := args[0].bulk

	SETsMu.RLock()
	value, ok := SETs[key]
	SETsMu.RUnlock()

	// ok: A boolean (true/false) that tells you whether the key existed
	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}

var HSETs = map[string]map[string]string{}
var HSETsMU = sync.RWMutex{}

func hset(args []Value) Value {
	if len(args) != 3 {
		return Value{
			typ: "error",
			str: "ERR wrong number of arguments for 'HSET' command",
		}
	}

	hash := args[0].bulk
	key := args[1].bulk
	value := args[2].bulk

	HSETsMU.Lock()
	// If this hash does NOT exist yet, create it.
	if _, ok := HSETs[hash]; !ok {
		HSETs[hash] = map[string]string{}
	}
	HSETs[hash][key] = value
	HSETsMU.Unlock()

	return Value{typ: "string", str: "OK"}
}

func hget(args []Value) Value {
	if len(args) != 2 {
		return Value{
			typ: "error",
			str: "ERR wrong number of arguments for 'HGET' command",
		}
	}

	hash := args[0].bulk
	key := args[1].bulk

	HSETsMU.RLock()
	value, ok := HSETs[hash][key]
	HSETsMU.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}

func hgetall(args []Value) Value {
	if len(args) != 1 {
		return Value{
			typ: "error",
			str: "ERR wrong number of arguments for 'HGETALL' command",
		}
	}

	hash := args[0].bulk

	HSETsMU.RLock()
	value, ok := HSETs[hash]
	HSETsMU.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	values := []Value{}
	for k, v := range value {
		values = append(values, Value{typ: "bulk", bulk: k})
		values = append(values, Value{typ: "bulk", bulk: v})
	}

	return Value{
		typ:   "array",
		array: values,
	}
}
