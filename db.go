package main

import (
	"sync"
)

type DB struct {
	mu  sync.RWMutex
	kv  map[string]string
	hkv map[string]map[string]string
}

func NewDB() *DB {
	return &DB{
		kv:  make(map[string]string),
		hkv: make(map[string]map[string]string),
	}
}

// String Operations

func (db *DB) Set(key, value string) {
	db.mu.Lock()
	// "Remember" to unlock at the very end
	defer db.mu.Unlock()

	db.kv[key] = value
}

func (db *DB) Get(key string) (string, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	val, ok := db.kv[key]

	return val, ok
}

// Hash Operation

func (db *DB) HSet(hash, key, value string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.hkv[hash]; !ok {
		db.hkv[hash] = make(map[string]string)
	}

	db.hkv[hash][key] = value
}

func (db *DB) HGet(hash, key string) (string, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if _, ok := db.hkv[hash]; !ok {
		return "", false
	}
	val, ok := db.hkv[hash][key]

	return val, ok
}

func (db *DB) HGetAll(hash string) (map[string]string, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	val, ok := db.hkv[hash]

	// Because you we the internal map directly,
	// the caller receives a reference to nil, which is safe to read
	return val, ok
}
