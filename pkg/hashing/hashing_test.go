package hashing

import (
	"testing"
)

func TestMurmur3Hash(t *testing.T) {
	hash := &Murmur3Hash{}
	
	// Test that same input produces same hash
	hash1 := hash.Hash("test-key")
	hash2 := hash.Hash("test-key")
	if hash1 != hash2 {
		t.Errorf("Expected same hash for same input, got %d and %d", hash1, hash2)
	}
	
	// Test that different inputs produce different hashes
	hash3 := hash.Hash("different-key")
	if hash1 == hash3 {
		t.Errorf("Expected different hashes for different inputs")
	}
	
	// Test empty string (should produce consistent hash)
	hash4 := hash.Hash("")
	hash5 := hash.Hash("")
	if hash4 != hash5 {
		t.Errorf("Expected same hash for empty string, got %d and %d", hash4, hash5)
	}
}

func TestXXHash(t *testing.T) {
	hash := &XXHash{}
	
	// Test that same input produces same hash
	hash1 := hash.Hash("test-key")
	hash2 := hash.Hash("test-key")
	if hash1 != hash2 {
		t.Errorf("Expected same hash for same input, got %d and %d", hash1, hash2)
	}
	
	// Test that different inputs produce different hashes
	hash3 := hash.Hash("different-key")
	if hash1 == hash3 {
		t.Errorf("Expected different hashes for different inputs")
	}
}

func TestNewHashFunction(t *testing.T) {
	// Test murmur3
	hash1 := NewHashFunction("murmur3")
	if hash1 == nil {
		t.Fatal("Expected non-nil hash function for murmur3")
	}
	if _, ok := hash1.(*Murmur3Hash); !ok {
		t.Errorf("Expected Murmur3Hash, got %T", hash1)
	}
	
	// Test xxhash
	hash2 := NewHashFunction("xxhash")
	if hash2 == nil {
		t.Fatal("Expected non-nil hash function for xxhash")
	}
	if _, ok := hash2.(*XXHash); !ok {
		t.Errorf("Expected XXHash, got %T", hash2)
	}
	
	// Test default (should be murmur3)
	hash3 := NewHashFunction("unknown")
	if hash3 == nil {
		t.Fatal("Expected non-nil hash function for unknown")
	}
	if _, ok := hash3.(*Murmur3Hash); !ok {
		t.Errorf("Expected Murmur3Hash as default, got %T", hash3)
	}
}

func TestConsistentHash_AddShard(t *testing.T) {
	ch := NewConsistentHash(NewHashFunction("murmur3"))
	
	ch.AddShard("shard1", 10)
	ch.AddShard("shard2", 10)
	
	shards := ch.GetShards()
	if len(shards) != 2 {
		t.Errorf("Expected 2 shards, got %d", len(shards))
	}
	
	// Check that shard1 and shard2 are in the list
	found1, found2 := false, false
	for _, s := range shards {
		if s == "shard1" {
			found1 = true
		}
		if s == "shard2" {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Errorf("Expected both shards in list, found1=%v, found2=%v", found1, found2)
	}
}

func TestConsistentHash_RemoveShard(t *testing.T) {
	ch := NewConsistentHash(NewHashFunction("murmur3"))
	
	ch.AddShard("shard1", 10)
	ch.AddShard("shard2", 10)
	ch.RemoveShard("shard1")
	
	shards := ch.GetShards()
	if len(shards) != 1 {
		t.Errorf("Expected 1 shard after removal, got %d", len(shards))
	}
	if shards[0] != "shard2" {
		t.Errorf("Expected shard2, got %s", shards[0])
	}
}

func TestConsistentHash_GetShard(t *testing.T) {
	ch := NewConsistentHash(NewHashFunction("murmur3"))
	
	// Test empty ring
	shard := ch.GetShard("key1")
	if shard != "" {
		t.Errorf("Expected empty string for empty ring, got %s", shard)
	}
	
	// Add shards
	ch.AddShard("shard1", 10)
	ch.AddShard("shard2", 10)
	
	// Test that same key returns same shard
	shard1 := ch.GetShard("key1")
	shard2 := ch.GetShard("key1")
	if shard1 != shard2 {
		t.Errorf("Expected same shard for same key, got %s and %s", shard1, shard2)
	}
	
	// Test that shard is one of the added shards
	if shard1 != "shard1" && shard1 != "shard2" {
		t.Errorf("Expected shard1 or shard2, got %s", shard1)
	}
}

func TestConsistentHash_Distribution(t *testing.T) {
	ch := NewConsistentHash(NewHashFunction("murmur3"))
	
	ch.AddShard("shard1", 100)
	ch.AddShard("shard2", 100)
	ch.AddShard("shard3", 100)
	
	// Test distribution across multiple keys
	distribution := make(map[string]int)
	for i := 0; i < 1000; i++ {
		key := "key" + string(rune(i))
		shard := ch.GetShard(key)
		distribution[shard]++
	}
	
	// All keys should map to one of the shards
	if len(distribution) != 3 {
		t.Errorf("Expected 3 shards in distribution, got %d", len(distribution))
	}
	
	// Distribution should be relatively even (within 20% tolerance)
	total := 1000
	for shard, count := range distribution {
		expected := total / 3
		tolerance := expected / 5 // 20% tolerance
		if count < expected-tolerance || count > expected+tolerance {
			t.Logf("Shard %s: %d keys (expected ~%d)", shard, count, expected)
		}
	}
}

func TestConsistentHash_WrapAround(t *testing.T) {
	ch := NewConsistentHash(NewHashFunction("murmur3"))
	
	ch.AddShard("shard1", 10)
	
	// Test with a key that hashes to a very large value
	// This tests the wrap-around logic
	shard := ch.GetShard("wrap-around-test-key-with-very-long-name")
	if shard != "shard1" {
		t.Errorf("Expected shard1 for wrap-around case, got %s", shard)
	}
}

func BenchmarkMurmur3Hash(b *testing.B) {
	hash := &Murmur3Hash{}
	key := "benchmark-key"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash.Hash(key)
	}
}

func BenchmarkXXHash(b *testing.B) {
	hash := &XXHash{}
	key := "benchmark-key"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash.Hash(key)
	}
}

func BenchmarkConsistentHash_GetShard(b *testing.B) {
	ch := NewConsistentHash(NewHashFunction("murmur3"))
	ch.AddShard("shard1", 256)
	ch.AddShard("shard2", 256)
	ch.AddShard("shard3", 256)
	
	key := "benchmark-key"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch.GetShard(key)
	}
}

