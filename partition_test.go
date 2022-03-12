package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeParitions(t *testing.T) {
	// 10 partitions with a keyrange from 1-10
	actual := MakePartitions(10, 10)
	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	assert.Equal(t, expected, actual, "Should've made 10 partitions from 1-10")

	// 10 partitions with keyrange 1-100
	actual = MakePartitions(100, 10)
	expected = []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	assert.Equal(t, expected, actual, "Should've made 10 partitions from 1-100")
}

func TestFindRange(t *testing.T) {
	partitions := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

	// 0 key falls in the first partition(0-10)
	expected := 0
	actual := FindRange(partitions, 0)
	assert.Equal(t, expected, actual, "Incorrect partition returned for key 0")

	// 100 key falls in last partition(90-100)
	expected = 9
	actual = FindRange(partitions, 100)
	assert.Equal(t, expected, actual, "Incorrect partition returned for key 100")

	// 50 key falls in the fourth partition(40-50)
	expected = 4
	actual = FindRange(partitions, 50)
	assert.Equal(t, expected, actual, "Incorrect partition returned for key 50")

}

// Not really sure how to test hash
func TestHash(t *testing.T) {
	key := "TestKey"
	maxRange := 1000

	hashVal := Hash(key, maxRange)
	assert.Less(t, hashVal, maxRange, "Hash value should be less than maxRange")
}
