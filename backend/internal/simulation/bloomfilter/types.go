package bloomfilter

// BloomFilterState represents the complete state of a Bloom Filter
type BloomFilterState struct {
	Size           int             `json:"size"`           // Size of bit array
	BitArray       []bool          `json:"bitArray"`       // The actual bit array
	NumHashFuncs   int             `json:"numHashFuncs"`   // Number of hash functions
	ItemsAdded     []string        `json:"itemsAdded"`     // Items that have been added
	FalsePositives int             `json:"falsePositives"` // Count of false positives detected
	TruePositives  int             `json:"truePositives"`  // Count of true positives
	TrueNegatives  int             `json:"trueNegatives"`  // Count of true negatives
	RecentOps      []Operation     `json:"recentOps"`      // Recent operations for visualization
}

// Operation represents a single operation on the Bloom Filter
type Operation struct {
	Type      string `json:"type"`      // "ADD" or "CHECK"
	Item      string `json:"item"`      // The item being operated on
	Result    string `json:"result"`    // For CHECK: "definitely_not", "probably_yes"
	ActualIn  bool   `json:"actualIn"`  // For CHECK: was it actually added?
	HashBits  []int  `json:"hashBits"`  // Bit positions affected by hash functions
	Timestamp string `json:"timestamp"` // When the operation occurred
}

// CheckResult represents the result of checking if an item exists
type CheckResult struct {
	Item            string `json:"item"`
	Result          string `json:"result"`          // "definitely_not" or "probably_yes"
	ActuallyPresent bool   `json:"actuallyPresent"` // Was it actually added?
	IsFalsePositive bool   `json:"isFalsePositive"` // True if probably_yes but not actually present
	HashBits        []int  `json:"hashBits"`        // Bit positions checked
}

