package bloomfilter

import (
	"hash/fnv"
	"sync"
	"time"
)

// BloomFilter represents a probabilistic data structure for set membership testing
type BloomFilter struct {
	mu             sync.RWMutex
	size           int      // Size of bit array
	bitArray       []bool   // The bit array
	numHashFuncs   int      // Number of hash functions to use
	itemsAdded     []string // Track items added (for simulation/visualization)
	falsePositives int      // Count of false positives
	truePositives  int      // Count of true positives
	trueNegatives  int      // Count of true negatives
	recentOps      []Operation
}

// NewBloomFilter creates a new Bloom Filter
// size: size of the bit array (larger = fewer false positives)
// numHashFuncs: number of hash functions (typically 3-5)
func NewBloomFilter(size, numHashFuncs int) *BloomFilter {
	return &BloomFilter{
		size:         size,
		bitArray:     make([]bool, size),
		numHashFuncs: numHashFuncs,
		itemsAdded:   []string{},
		recentOps:    []Operation{},
	}
}

// hash generates a hash value for the given item and seed
func (bf *BloomFilter) hash(item string, seed int) int {
	h := fnv.New32a()
	h.Write([]byte(item))
	h.Write([]byte{byte(seed)})
	hashValue := int(h.Sum32())
	if hashValue < 0 {
		hashValue = -hashValue
	}
	return hashValue % bf.size
}

// getHashPositions returns all hash positions for an item
func (bf *BloomFilter) getHashPositions(item string) []int {
	positions := make([]int, bf.numHashFuncs)
	for i := 0; i < bf.numHashFuncs; i++ {
		positions[i] = bf.hash(item, i)
	}
	return positions
}

// Add adds an item to the Bloom Filter
func (bf *BloomFilter) Add(item string) []int {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	positions := bf.getHashPositions(item)
	
	// Set all corresponding bits to 1
	for _, pos := range positions {
		bf.bitArray[pos] = true
	}

	// Track the item (for simulation purposes)
	bf.itemsAdded = append(bf.itemsAdded, item)

	// Record operation
	op := Operation{
		Type:      "ADD",
		Item:      item,
		HashBits:  positions,
		Timestamp: time.Now().Format("15:04:05"),
	}
	bf.recentOps = append(bf.recentOps, op)
	
	// Keep only last 10 operations
	if len(bf.recentOps) > 10 {
		bf.recentOps = bf.recentOps[len(bf.recentOps)-10:]
	}

	return positions
}

// Check checks if an item might be in the set
// Returns: "definitely_not" or "probably_yes"
func (bf *BloomFilter) Check(item string) CheckResult {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	positions := bf.getHashPositions(item)
	
	// Check if all corresponding bits are set
	allSet := true
	for _, pos := range positions {
		if !bf.bitArray[pos] {
			allSet = false
			break
		}
	}

	// Check if item was actually added
	actuallyPresent := false
	for _, added := range bf.itemsAdded {
		if added == item {
			actuallyPresent = true
			break
		}
	}

	result := "definitely_not"
	isFalsePositive := false

	if allSet {
		result = "probably_yes"
		if !actuallyPresent {
			isFalsePositive = true
			bf.falsePositives++
		} else {
			bf.truePositives++
		}
	} else {
		bf.trueNegatives++
	}

	// Record operation
	op := Operation{
		Type:      "CHECK",
		Item:      item,
		Result:    result,
		ActualIn:  actuallyPresent,
		HashBits:  positions,
		Timestamp: time.Now().Format("15:04:05"),
	}
	bf.recentOps = append(bf.recentOps, op)
	
	// Keep only last 10 operations
	if len(bf.recentOps) > 10 {
		bf.recentOps = bf.recentOps[len(bf.recentOps)-10:]
	}

	return CheckResult{
		Item:            item,
		Result:          result,
		ActuallyPresent: actuallyPresent,
		IsFalsePositive: isFalsePositive,
		HashBits:        positions,
	}
}

// GetState returns the current state of the Bloom Filter
func (bf *BloomFilter) GetState() BloomFilterState {
	bf.mu.RLock()
	defer bf.mu.RUnlock()

	return BloomFilterState{
		Size:           bf.size,
		BitArray:       bf.bitArray,
		NumHashFuncs:   bf.numHashFuncs,
		ItemsAdded:     bf.itemsAdded,
		FalsePositives: bf.falsePositives,
		TruePositives:  bf.truePositives,
		TrueNegatives:  bf.trueNegatives,
		RecentOps:      bf.recentOps,
	}
}

// Reset resets the Bloom Filter to initial state
func (bf *BloomFilter) Reset() {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	bf.bitArray = make([]bool, bf.size)
	bf.itemsAdded = []string{}
	bf.falsePositives = 0
	bf.truePositives = 0
	bf.trueNegatives = 0
	bf.recentOps = []Operation{}
}

// GetFillPercentage returns the percentage of bits set to 1
func (bf *BloomFilter) GetFillPercentage() float64 {
	bf.mu.RLock()
	defer bf.mu.RUnlock()

	count := 0
	for _, bit := range bf.bitArray {
		if bit {
			count++
		}
	}
	return float64(count) / float64(bf.size) * 100.0
}

