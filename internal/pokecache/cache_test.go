package pokecache

import (
	"testing"
	"time"
)

func TestAdd(t *testing.T) {
	cache := NewCache(5 * time.Second)
	key := "test-key"
	value := []byte("test-value")

	cache.Add(key, value)

	// Verify the value was added
	retrieved, ok := cache.Get(key)
	if !ok {
		t.Errorf("Expected to find key %q in cache, but it was not found", key)
	}
	if string(retrieved) != string(value) {
		t.Errorf("Expected value %q, got %q", string(value), string(retrieved))
	}
}

func TestGet(t *testing.T) {
	cache := NewCache(5 * time.Second)
	key := "test-key"
	value := []byte("test-value")

	// Test getting a non-existent key
	_, ok := cache.Get("non-existent")
	if ok {
		t.Error("Expected false for non-existent key, got true")
	}

	// Add a value and test getting it
	cache.Add(key, value)
	retrieved, ok := cache.Get(key)
	if !ok {
		t.Errorf("Expected to find key %q in cache, but it was not found", key)
	}
	if string(retrieved) != string(value) {
		t.Errorf("Expected value %q, got %q", string(value), string(retrieved))
	}
}

func TestReapLoop(t *testing.T) {
	// Use a very short interval for testing (100ms)
	interval := 100 * time.Millisecond
	cache := NewCache(interval)
	key := "test-key"
	value := []byte("test-value")

	// Add a value
	cache.Add(key, value)

	// Verify it's there
	_, ok := cache.Get(key)
	if !ok {
		t.Error("Expected to find key immediately after adding, but it was not found")
	}

	// Wait for the entry to expire
	time.Sleep(interval + 50*time.Millisecond)

	// Verify it's been reaped (removed)
	_, ok = cache.Get(key)
	if ok {
		t.Error("Expected key to be reaped after interval, but it was still found")
	}
}
