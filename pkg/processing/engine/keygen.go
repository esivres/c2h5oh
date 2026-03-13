package engine

import "sync/atomic"

// partitionBits is the number of high bits reserved for the partition id.
const partitionBits = 56

// KeyGenerator produces unique uint64 keys with partition prefix in the high 8 bits.
//
// Layout: | partition (8 bit) | sequence (56 bit) |
//
// This allows O(1) partition lookup from any key.
type KeyGenerator struct {
	prefix   uint64 // partitionId shifted to high 8 bits
	sequence atomic.Uint64
}

// NewKeyGenerator creates a key generator for a given partition.
func NewKeyGenerator(partitionId uint8, startSequence uint64) *KeyGenerator {
	kg := &KeyGenerator{
		prefix: uint64(partitionId) << partitionBits,
	}
	kg.sequence.Store(startSequence)
	return kg
}

// Next returns the next unique key.
func (kg *KeyGenerator) Next() uint64 {
	seq := kg.sequence.Add(1)
	return kg.prefix | seq
}

// PartitionOf extracts the partition id from a key.
func PartitionOf(key uint64) uint8 {
	return uint8(key >> partitionBits)
}
