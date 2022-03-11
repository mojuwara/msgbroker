package main

import (
	"crypto/sha256"
	"encoding/binary"
)

// n is the key range(from [0, n], consider having a start?)
// p is the number of partitions you want
// Does not support floating point numbers
// Returns an slice of numbers, each number in the array is where that range ends
func MakePartitions(keyRange int, partitions int) []int {
	rangeSize := keyRange / partitions

	i := 1
	result := []int{}
	for i <= partitions {
		rangeEnd := rangeSize * i
		result = append(result, rangeEnd)
		i++
	}
	return result
}

// SHA256 the given key and mod it so it falls in key range [0, r)
// TODO: Cache hash values
func Hash(key string, maxRange int) int {
	hasher := sha256.New()
	hasher.Write([]byte(key))
	endianVal := binary.LittleEndian.Uint64(hasher.Sum(nil))
	return int(endianVal) % maxRange
}

// Returns the index of the range it falls into, or -1 if not found
// TODO: Improve by using binary search
func FindRange(parts []int, desired int) int {
	for ndx, rangeEnd := range parts {
		if desired <= rangeEnd {
			return ndx
		}
	}
	return -1
}
