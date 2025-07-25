package datastructures

import (
	"hash"
	"hash/fnv"
	"math"
	"math/big"
)

type BloomFilter struct {
	size   int
	hashes []hash.Hash64
	bits   *big.Int
}

func NewBloomFilter(size int, hashCount int) *BloomFilter {
	bits := &big.Int{}
	bits.FillBytes(make([]byte, (size+8-1)/8))
	hashes := make([]hash.Hash64, hashCount)
	for i := range hashCount {
		hashes[i] = fnv.New64()
	}
	return &BloomFilter{
		bits:   bits,
		hashes: hashes,
		size:   size,
	}
}

func (filter *BloomFilter) Add(item string) {
	for _, hashFunc := range filter.hashes {
		hashFunc.Reset()
		hashFunc.Write([]byte(item))
		index := hashFunc.Sum64() % uint64(filter.size)
		filter.bits = filter.bits.SetBit(filter.bits, int(index), 1)
	}
}

func (filter *BloomFilter) Check(item string) bool {
	for _, hashFunc := range filter.hashes {
		hashFunc.Reset()
		hashFunc.Write([]byte(item))
		index := hashFunc.Sum64() % uint64(filter.size)
		if filter.bits.Bit(int(index)) == 0 {
			return false
		}
	}
	return true
}

func NewOptimalBloomFilter(numElements int, falsePositiveRate float64) *BloomFilter {
	m, k := optimalParams(numElements, falsePositiveRate)
	return NewBloomFilter(m, k)
}

func optimalParams(n int, p float64) (int, int) {
	m := int(math.Ceil(float64(-n) * math.Log(p) / (math.Pow(math.Log(2), 2))))
	k := int(math.Ceil(math.Log(2) * float64(m) / float64(n)))
	return m, k
}
