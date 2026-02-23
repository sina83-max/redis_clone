package main

// a map whose keys are strings, and whose values are functions.
var Handlers = map[string]func(args []Value, db *DB) Value{
	"PING":    ping,
	"SET":     set,
	"GET":     get,
	"HSET":    hset,
	"HGET":    hget,
	"HGETALL": hgetall,
}

func ping(args []Value, db *DB) Value {
	if len(args) == 0 {
		return Value{typ: "string", str: "Pong"}
	}

	return Value{typ: "string", str: args[0].bulk}
}

func set(args []Value, db *DB) Value {
	if len(args) != 2 {
		return Value{
			typ: "error",
			str: "ERR wrong number of arguments for 'SET' command",
		}
	}

	db.Set(args[0].bulk, args[1].bulk)

	return Value{typ: "string", str: "OK"}
}

func get(args []Value, db *DB) Value {
	if len(args) != 1 {
		return Value{
			typ: "error",
			str: "ERR wrong number of arguments for 'GET' command",
		}
	}

	value, ok := db.Get(args[0].bulk)

	// ok: A boolean (true/false) that tells you whether the key existed
	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}

func hset(args []Value, db *DB) Value {
	if len(args) != 3 {
		return Value{
			typ: "error",
			str: "ERR wrong number of arguments for 'HSET' command",
		}
	}

	hash := args[0].bulk
	key := args[1].bulk
	value := args[2].bulk

	db.HSet(hash, key, value)

	return Value{typ: "string", str: "OK"}
}

func hget(args []Value, db *DB) Value {
	if len(args) != 2 {
		return Value{
			typ: "error",
			str: "ERR wrong number of arguments for 'HGET' command",
		}
	}

	hash := args[0].bulk
	key := args[1].bulk

	value, ok := db.HGet(hash, key)
	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}

func hgetall(args []Value, db *DB) Value {
	if len(args) != 1 {
		return Value{
			typ: "error",
			str: "ERR wrong number of arguments for 'HGETALL' command",
		}
	}

	hash := args[0].bulk

	data, ok := db.HGetAll(hash)

	if !ok {
		return Value{typ: "null"}
	}

	values := []Value{}
	for k, v := range data {
		values = append(values, Value{typ: "bulk", bulk: k})
		values = append(values, Value{typ: "bulk", bulk: v})
	}

	return Value{
		typ:   "array",
		array: values,
	}
}
