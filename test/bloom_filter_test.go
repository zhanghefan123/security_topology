package test

import (
	"encoding/binary"
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"testing"
)

func TestBloomFilter(t *testing.T) {
	M := uint(200)
	K := uint(1)
	bloomFilter := bloom.New(M, K)
	numberOfElements := uint32(10)
	for i := range numberOfElements {
		bytes := make([]byte, 4)
		binary.BigEndian.PutUint32(bytes, i)
		bloomFilter.Add(bytes)
	}
	// output the false positive rate
	for i := range numberOfElements {
		bytes := make([]byte, 4)
		binary.BigEndian.PutUint32(bytes, i)
		result := bloomFilter.Test(bytes)
		if result {
			fmt.Printf("element %d exists\n", i)
		} else {
			fmt.Printf("element %d does not exists\n", i)
		}
	}
	// test the false positive rate of the bloom filter
	numberOfBitsSet := bloomFilter.BitSet().Count()
	// test the number of inserted elements
	numberOfInsertedElements := bloomFilter.ApproximatedSize()
	// output the size
	fmt.Printf("number of bits %d , number of inserted elements: %d\n", numberOfBitsSet, numberOfInsertedElements)
}
