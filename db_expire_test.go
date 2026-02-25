package main

import (
	"testing"
	"time"
)

func TestExpireAndTTL(t *testing.T) {
	db := NewDB()
	db.Set("k1", "v1")

	// set TTL to 1 second
	if !db.Expire("k1", 5) {
		t.Fatalf("expected Expire to return true")
	}

	ttl := db.TTL("k1")
	if ttl <= 0 {
		t.Fatalf("expected positive TTL, got %d", ttl)
	}

	// wait for expiration
	time.Sleep(5100 * time.Millisecond)

	// key should be gone
	if _, ok := db.Get("k1"); ok {
		t.Fatalf("expected key to be expired and deleted")
	}
	if db.TTL("k1") != -2 {
		t.Fatalf("expected TTL -2 for non-existent key")
	}
}

func TestHSetAfterExpire(t *testing.T) {
	db := NewDB()
	db.HSet("h1", "field1", "v1")
	db.Expire("h1", 1)
	time.Sleep(1100 * time.Millisecond)

	// HSet on expired hash should create it again
	db.HSet("h1", "field2", "v2")
	if v, ok := db.HGet("h1", "field2"); !ok || v != "v2" {
		t.Fatalf("expected new hash to be created after expire")
	}
}
