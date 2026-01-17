package store

import (
	"sync"
	"testing"
	"time"
)

func TestMemoryStore_GetSet(t *testing.T) {
	store := NewMemoryStore(0)
	defer store.Close()

	bucket := &Bucket{
		Tokens:       10,
		LastRefillAt: time.Now(),
	}

	// Set a bucket
	store.Set("test-key", bucket)

	// Get the bucket
	retrieved := store.Get("test-key")
	if retrieved == nil {
		t.Fatal("Expected bucket to exist")
	}

	if retrieved.Tokens != bucket.Tokens {
		t.Errorf("Expected %d tokens, got %d", bucket.Tokens, retrieved.Tokens)
	}
}

func TestMemoryStore_GetNonExistent(t *testing.T) {
	store := NewMemoryStore(0)
	defer store.Close()

	bucket := store.Get("non-existent")
	if bucket != nil {
		t.Error("Expected nil for non-existent key")
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	store := NewMemoryStore(0)
	defer store.Close()

	bucket := &Bucket{
		Tokens:       10,
		LastRefillAt: time.Now(),
	}

	store.Set("test-key", bucket)
	store.Delete("test-key")

	retrieved := store.Get("test-key")
	if retrieved != nil {
		t.Error("Expected bucket to be deleted")
	}
}

func TestMemoryStore_Clear(t *testing.T) {
	store := NewMemoryStore(0)
	defer store.Close()

	// Add multiple buckets
	for i := 0; i < 10; i++ {
		bucket := &Bucket{
			Tokens:       int64(i),
			LastRefillAt: time.Now(),
		}
		store.Set(string(rune('A'+i)), bucket)
	}

	if store.Size() != 10 {
		t.Errorf("Expected 10 buckets, got %d", store.Size())
	}

	// Clear all
	store.Clear()

	if store.Size() != 0 {
		t.Errorf("Expected 0 buckets after clear, got %d", store.Size())
	}
}

func TestMemoryStore_Concurrency(t *testing.T) {
	store := NewMemoryStore(0)
	defer store.Close()

	const numGoroutines = 100
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				bucket := &Bucket{
					Tokens:       int64(j),
					LastRefillAt: time.Now(),
				}
				key := string(rune('A' + (id % 26)))
				store.Set(key, bucket)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := string(rune('A' + (id % 26)))
				store.Get(key)
			}
		}(i)
	}

	wg.Wait()
}

func TestMemoryStore_Cleanup(t *testing.T) {
	// Create store with short cleanup interval
	store := NewMemoryStore(100 * time.Millisecond)
	defer store.Close()

	// Add a bucket with old timestamp
	oldBucket := &Bucket{
		Tokens:       10,
		LastRefillAt: time.Now().Add(-2 * time.Hour),
	}
	store.Set("old-key", oldBucket)

	// Add a bucket with recent timestamp
	newBucket := &Bucket{
		Tokens:       10,
		LastRefillAt: time.Now(),
	}
	store.Set("new-key", newBucket)

	if store.Size() != 2 {
		t.Errorf("Expected 2 buckets, got %d", store.Size())
	}

	// Wait for cleanup to run
	time.Sleep(200 * time.Millisecond)

	// Old bucket should be removed, new bucket should remain
	if store.Size() != 1 {
		t.Errorf("Expected 1 bucket after cleanup, got %d", store.Size())
	}

	if store.Get("old-key") != nil {
		t.Error("Old bucket should be removed")
	}

	if store.Get("new-key") == nil {
		t.Error("New bucket should still exist")
	}
}

func TestMemoryStore_IsolatedCopies(t *testing.T) {
	store := NewMemoryStore(0)
	defer store.Close()

	original := &Bucket{
		Tokens:       10,
		LastRefillAt: time.Now(),
	}

	store.Set("test-key", original)

	// Modify the original
	original.Tokens = 999

	// Retrieved bucket should not be affected
	retrieved := store.Get("test-key")
	if retrieved.Tokens != 10 {
		t.Errorf("Expected 10 tokens (isolated copy), got %d", retrieved.Tokens)
	}

	// Modify retrieved bucket
	retrieved.Tokens = 888

	// Store should not be affected
	retrieved2 := store.Get("test-key")
	if retrieved2.Tokens != 10 {
		t.Errorf("Expected 10 tokens (isolated copy), got %d", retrieved2.Tokens)
	}
}

func BenchmarkMemoryStore_Set(b *testing.B) {
	store := NewMemoryStore(0)
	defer store.Close()

	bucket := &Bucket{
		Tokens:       10,
		LastRefillAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Set("bench-key", bucket)
	}
}

func BenchmarkMemoryStore_Get(b *testing.B) {
	store := NewMemoryStore(0)
	defer store.Close()

	bucket := &Bucket{
		Tokens:       10,
		LastRefillAt: time.Now(),
	}
	store.Set("bench-key", bucket)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Get("bench-key")
	}
}

func BenchmarkMemoryStore_Parallel(b *testing.B) {
	store := NewMemoryStore(0)
	defer store.Close()

	bucket := &Bucket{
		Tokens:       10,
		LastRefillAt: time.Now(),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				store.Set("bench-key", bucket)
			} else {
				store.Get("bench-key")
			}
			i++
		}
	})
}
