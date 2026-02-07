# Redis Clone - In-Memory Database with AOF Persistence

A lightweight Redis-compatible in-memory database implemented in Go, featuring RESP protocol support and Append-Only File persistence.

## Features

- Redis Protocol Compatibility: Fully implements RESP (Redis Serialization Protocol)
- In-Memory Storage: Fast key-value operations with hash table backing
- AOF Persistence: Append-Only File persistence for data durability
- Concurrent Safe: Thread-safe operations with mutex protection
