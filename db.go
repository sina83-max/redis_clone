package main

import (
	"sync"
	"time"
)

type DB struct {
	mu      sync.RWMutex
	kv      map[string]string
	hkv     map[string]map[string]string
	expires map[string]time.Time
}

func NewDB() *DB {
	return &DB{
		kv:  make(map[string]string),
		hkv: make(map[string]map[string]string),
		// absolute time of expiration - not duration
		expires: make(map[string]time.Time),
	}
}

// Expiration API

// Sets an absolute expiratin time
// key exists and time set -> True
// Key exists but alredy expired -> False
// Not exists -> False
func (db *DB) Expire(key string, seconds int) bool {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Check Existance
	_, inKV := db.kv[key]
	_, inHKV := db.hkv[key]
	if !inKV && !inHKV {
		return false
	}

	// Check if already expired
	if exp, ok := db.expires[key]; ok {
		if time.Now().After(exp) {
			// Key is logically dead
			delete(db.kv, key)
			delete(db.hkv, key)
			delete(db.expires, key)

			return false
		}
	}

	// Set new expiration
	db.expires[key] = time.Now().Add(time.Duration(seconds) * time.Second)
	return true

}

// TTL returns remaining TTL in seconds.
// Follows Redis semantics:
// - return -2 if the key does not exist.
// - return -1 if the key exists but has no associated expire.
func (db *DB) TTL(key string) int {
	db.mu.RLock()
	// Because we are going to change the
	// Lock type, we don't use "defer"

	// Check Existance
	_, inKV := db.kv[key]
	_, inHKV := db.hkv[key]
	if !inKV && !inHKV {
		db.mu.RUnlock()
		return -2
	}

	exp, hasExpirationKey := db.expires[key]
	if !hasExpirationKey {
		db.mu.RUnlock()
		return -1
	}

	// if expired -> delete and return -2
	if time.Now().After(exp) {
		db.mu.RUnlock()
		// acquire full lock
		// Recheck to avoid race in the meantime
		db.mu.Lock()
		// recheck
		exp2, hasExpirationKey2 := db.expires[key]
		if hasExpirationKey2 && time.Now().After(exp2) {
			// Because of manually controlling the lock
			// didn't use deleteKey()
			delete(db.kv, key)
			delete(db.hkv, key)
			delete(db.expires, key)
		}

		db.mu.Unlock()
		return -2
	}

	remaining := int(time.Until(exp).Seconds())
	db.mu.RUnlock()
	if remaining < 0 {
		return -2
	}
	return remaining
}

// isExpired checks whether key has an expiration that is already due.
// IMPORTANT: isExpired does NOT delete the key;
// callers should delete if needed.
func (db *DB) isExpired(key string) bool {
	db.mu.RLock()
	exp, ok := db.expires[key]
	db.mu.RUnlock()
	if !ok {
		return false
	}

	return time.Now().After(exp)
}

// deleteKey removes a key from kv/hkv and expires maps.
// It acquires the write-lock internally.
func (db *DB) deleteKey(key string) {
	delete(db.kv, key)
	delete(db.hkv, key)
	delete(db.expires, key)
}

// String Operations

func (db *DB) Set(key, value string) {
	db.mu.Lock()
	// SET should overwrite and remove
	// any previous expiry (same as Redis).
	delete(db.expires, key)
	db.kv[key] = value
	db.mu.Unlock()
}

func (db *DB) Get(key string) (string, bool) {
	db.mu.RLock()
	val, ok := db.kv[key]
	if !ok {
		db.mu.RUnlock()
		return "", false
	}

	// check expiration under RLock
	exp, hasExp := db.expires[key]
	if hasExp && time.Now().After(exp) {
		// expired — must delete under write lock after re-check
		db.mu.RUnlock()
		db.mu.Lock()
		// re-check to avoid race with other writers
		exp2, has2 := db.expires[key]
		if has2 && time.Now().After(exp2) {
			db.deleteKey(key)
		}
		db.mu.Unlock()
		return "", false
	}

	// not expired; return value
	db.mu.RUnlock()
	return val, true
}

// Hash Operation

func (db *DB) HSet(hash, key, value string) {
	db.mu.Lock()
	// If the hash itself was expired -
	// treat it as non-existing;
	// remove old data and expire time
	if exp, has := db.expires[hash]; has && time.Now().After(exp) {
		db.deleteKey(hash)
	}
	// Ensure hash exists
	if _, ok := db.hkv[hash]; !ok {
		db.hkv[hash] = make(map[string]string)
	}
	db.hkv[hash][key] = value
	db.mu.Unlock()
}

func (db *DB) HGet(hash, key string) (string, bool) {
	db.mu.RLock()
	hmap, ok := db.hkv[hash]
	if !ok {
		db.mu.RUnlock()
		return "", false
	}

	// Check if hash is expired
	if exp, has := db.expires[hash]; has && time.Now().After(exp) {
		db.mu.RUnlock()
		// delete after full-lock and
		// re-check for race condition
		db.mu.Lock()
		if exp2, has2 := db.expires[hash]; has2 && time.Now().After(exp2) {
			db.deleteKey(hash)
		}
		db.mu.Unlock()
		return "", false
	}

	// Safe to read field
	val, ok := hmap[key]
	db.mu.RUnlock()
	if !ok {
		return "", false
	}

	return val, true
}

func (db *DB) HGetAll(hash string) (map[string]string, bool) {
	db.mu.RLock()
	hmap, ok := db.hkv[hash]
	if !ok {
		db.mu.RUnlock()
		return nil, false
	}

	if exp, has := db.expires[hash]; has && time.Now().After(exp) {
		db.mu.RUnlock()
		db.mu.Lock()
		if exp2, has2 := db.expires[hash]; has2 && time.Now().After(exp2) {
			db.deleteKey(hash)
		}
		db.mu.Unlock()
		return nil, false
	}

	// Return a copy to avoid exposing the original
	// map to callers
	// out := make(map[string]string, len(hmap))
	// for k, v := range out {
	// 	out[k] = v
	// }
	// db.mu.RUnlock()
	// return out, true

	return hmap, true
}
