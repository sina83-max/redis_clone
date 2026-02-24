package main

import "testing"

func TestDB_SetGet(t *testing.T) {
	db := NewDB()

	// Test Set and Get
	db.Set("name", "sina")
	val, ok := db.Get("name")

	if !ok {
		t.Errorf("Expected to find key 'name'")
	}
	if val != "sina" {
		t.Errorf("Expected to get 'sina', get %s", val)
	}

	// Test missing key
	_, ok = db.Get("nonexistant")
	if ok {
		t.Errorf("Expecetd 'ok' to be false for a non-existant key, but It's true")
	}
}

func TestDB_Hashes(t *testing.T) {
	db := NewDB()

	// Test HSet and HGet
	db.HSet("details", "name", "sina")
	val, ok := db.HGet("details", "name")

	if !ok || val != "sina" {
		t.Errorf("HSet/HGet failed: got %s, %v", val, ok)
	}

	// Test HGetAll
	data, ok := db.HGetAll("details")
	if !ok || len(data) != 1 {
		t.Errorf("HGetAll failed to retrieve correct hash data")
	}
}
