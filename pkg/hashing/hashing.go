package hashing

import (
	"github.com/cespare/xxhash/v2"
	"github.com/spaolacci/murmur3"
)

// HashFunction defines the interface for hash functions
type HashFunction interface {
	Hash(key string) uint64
}

// Murmur3Hash implements Murmur3 hash
type Murmur3Hash struct{}

func (m *Murmur3Hash) Hash(key string) uint64 {
	h := murmur3.New64()
	h.Write([]byte(key))
	return h.Sum64()
}

// XXHash implements xxHash
type XXHash struct{}

func (x *XXHash) Hash(key string) uint64 {
	return xxhash.Sum64String(key)
}

// NewHashFunction creates a hash function based on name
func NewHashFunction(name string) HashFunction {
	switch name {
	case "xxhash":
		return &XXHash{}
	case "murmur3":
		fallthrough
	default:
		return &Murmur3Hash{}
	}
}

// ConsistentHash implements consistent hashing with virtual nodes
type ConsistentHash struct {
	hashFunc HashFunction
	vnodes   []VNodeEntry
}

// VNodeEntry represents a virtual node entry
type VNodeEntry struct {
	Hash   uint64
	ShardID string
}

// NewConsistentHash creates a new consistent hash ring
func NewConsistentHash(hashFunc HashFunction) *ConsistentHash {
	return &ConsistentHash{
		hashFunc: hashFunc,
		vnodes:   make([]VNodeEntry, 0),
	}
}

// AddShard adds a shard with virtual nodes to the ring
func (ch *ConsistentHash) AddShard(shardID string, vnodeCount int) {
	for i := 0; i < vnodeCount; i++ {
		vnodeKey := shardID + "-vnode-" + string(rune(i))
		hash := ch.hashFunc.Hash(vnodeKey)
		ch.vnodes = append(ch.vnodes, VNodeEntry{
			Hash:    hash,
			ShardID: shardID,
		})
	}
	ch.sortVNodes()
}

// RemoveShard removes a shard and its virtual nodes from the ring
func (ch *ConsistentHash) RemoveShard(shardID string) {
	newVNodes := make([]VNodeEntry, 0)
	for _, vnode := range ch.vnodes {
		if vnode.ShardID != shardID {
			newVNodes = append(newVNodes, vnode)
		}
	}
	ch.vnodes = newVNodes
	ch.sortVNodes()
}

// GetShard returns the shard ID for a given key
func (ch *ConsistentHash) GetShard(key string) string {
	if len(ch.vnodes) == 0 {
		return ""
	}

	keyHash := ch.hashFunc.Hash(key)
	
	// Binary search for the first vnode with hash >= keyHash
	idx := ch.findVNode(keyHash)
	return ch.vnodes[idx].ShardID
}

// findVNode finds the vnode for a given hash using binary search
func (ch *ConsistentHash) findVNode(hash uint64) int {
	left, right := 0, len(ch.vnodes)
	
	for left < right {
		mid := (left + right) / 2
		if ch.vnodes[mid].Hash < hash {
			left = mid + 1
		} else {
			right = mid
		}
	}
	
	// Wrap around if hash is greater than all vnodes
	if left >= len(ch.vnodes) {
		left = 0
	}
	
	return left
}

// sortVNodes sorts virtual nodes by hash value
func (ch *ConsistentHash) sortVNodes() {
	// Simple insertion sort for small-medium sized arrays
	// For larger arrays, consider using sort.Slice
	for i := 1; i < len(ch.vnodes); i++ {
		key := ch.vnodes[i]
		j := i - 1
		for j >= 0 && ch.vnodes[j].Hash > key.Hash {
			ch.vnodes[j+1] = ch.vnodes[j]
			j--
		}
		ch.vnodes[j+1] = key
	}
}

// GetShards returns all shard IDs in the ring
func (ch *ConsistentHash) GetShards() []string {
	shardMap := make(map[string]bool)
	for _, vnode := range ch.vnodes {
		shardMap[vnode.ShardID] = true
	}
	
	shards := make([]string, 0, len(shardMap))
	for shardID := range shardMap {
		shards = append(shards, shardID)
	}
	return shards
}

