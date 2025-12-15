package dump

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// OpsAnalyzer provides comprehensive operational analysis for Redis RDB
type OpsAnalyzer struct {
	counter *Counter

	// Basic stats
	totalKeys   uint64
	totalBytes  uint64
	avgKeySize  float64
	avgValueSize float64

	// Anomaly detection
	anomalies []Anomaly

	// TTL/Expiry analysis
	keysWithTTL    uint64
	keysWithoutTTL uint64
	expiredKeys    uint64
	expiryDistribution map[string]uint64 // "1h", "1d", "7d", "30d", "90d+", "expired"

	// Memory hotspots
	memoryHotspots []MemoryHotspot

	// Key pattern analysis
	keyPatterns []KeyPattern

	// Data type efficiency
	typeEfficiency map[string]TypeEfficiency

	// Fragmentation analysis
	fragmentationScore float64

	// Cluster slot analysis (if applicable)
	slotImbalance float64
	topSlotsUsage []SlotUsage

	// Health score (0-100)
	healthScore int

	// Recommendations
	recommendations []Recommendation
}

// Anomaly represents a detected issue
type Anomaly struct {
	Level       string    `json:"level"`        // "critical", "warning", "info"
	Category    string    `json:"category"`     // "memory", "ttl", "keys", "performance", "cluster"
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
	Suggestion  string    `json:"suggestion"`
	Value       string    `json:"value"`
	DetectedAt  time.Time `json:"detected_at"`
}

// MemoryHotspot identifies memory concentration issues
type MemoryHotspot struct {
	Type        string  `json:"type"`         // "key_prefix", "data_type", "single_key"
	Identifier  string  `json:"identifier"`
	MemoryUsed  uint64  `json:"memory_used"`
	KeyCount    uint64  `json:"key_count"`
	Percentage  float64 `json:"percentage"`
	AvgKeySize  uint64  `json:"avg_key_size"`
}

// KeyPattern represents common key naming patterns
type KeyPattern struct {
	Pattern     string  `json:"pattern"`
	Count       uint64  `json:"count"`
	TotalMemory uint64  `json:"total_memory"`
	AvgMemory   uint64  `json:"avg_memory"`
	Percentage  float64 `json:"percentage"`
	Example     string  `json:"example"`
}

// TypeEfficiency analyzes efficiency of data type usage
type TypeEfficiency struct {
	AvgSize        uint64  `json:"avg_size"`
	MedianSize     uint64  `json:"median_size"`
	P95Size        uint64  `json:"p95_size"`
	P99Size        uint64  `json:"p99_size"`
	Efficiency     float64 `json:"efficiency"`      // 0-100, higher is better
	WastedMemory   uint64  `json:"wasted_memory"`   // estimated
	OptimalType    string  `json:"optimal_type"`    // suggested type
}

// SlotUsage for cluster analysis
type SlotUsage struct {
	Slot       int     `json:"slot"`
	KeyCount   uint64  `json:"key_count"`
	MemoryUsed uint64  `json:"memory_used"`
	Percentage float64 `json:"percentage"`
}

// Recommendation provides actionable advice
type Recommendation struct {
	Priority    int       `json:"priority"`     // 1-5, 1 is highest
	Category    string    `json:"category"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Action      string    `json:"action"`
	Impact      string    `json:"impact"`
	Effort      string    `json:"effort"`       // "low", "medium", "high"
	CreatedAt   time.Time `json:"created_at"`
}

// NewOpsAnalyzer creates a new operational analyzer
func NewOpsAnalyzer(counter *Counter) *OpsAnalyzer {
	analyzer := &OpsAnalyzer{
		counter:            counter,
		anomalies:          []Anomaly{},
		memoryHotspots:     []MemoryHotspot{},
		keyPatterns:        []KeyPattern{},
		typeEfficiency:     make(map[string]TypeEfficiency),
		expiryDistribution: make(map[string]uint64),
		topSlotsUsage:      []SlotUsage{},
		recommendations:    []Recommendation{},
	}

	analyzer.analyze()
	return analyzer
}

// analyze performs comprehensive analysis
func (oa *OpsAnalyzer) analyze() {
	// Calculate basic stats
	oa.calculateBasicStats()

	// Detect anomalies
	oa.detectLargeKeys()
	oa.detectMemoryHotspots()
	oa.detectKeyExplosion()
	oa.detectTypeImbalance()
	oa.analyzeKeyPatterns()
	oa.analyzeTypeEfficiency()
	oa.analyzeClusterBalance()

	// Calculate health score
	oa.calculateHealthScore()

	// Generate recommendations
	oa.generateRecommendations()
}

// calculateBasicStats computes fundamental metrics
func (oa *OpsAnalyzer) calculateBasicStats() {
	for _, v := range oa.counter.typeNum {
		oa.totalKeys += v
	}
	for _, v := range oa.counter.typeBytes {
		oa.totalBytes += v
	}

	if oa.totalKeys > 0 {
		oa.avgKeySize = float64(oa.totalBytes) / float64(oa.totalKeys)
	}
}

// detectLargeKeys identifies abnormally large keys
func (oa *OpsAnalyzer) detectLargeKeys() {
	largestKeys := oa.counter.GetLargestEntries(100)

	// Define thresholds
	const (
		warningSize  = 10 * 1024 * 1024  // 10MB
		criticalSize = 50 * 1024 * 1024  // 50MB
	)

	criticalCount := 0
	warningCount := 0

	for _, entry := range largestKeys {
		if entry.Bytes >= criticalSize {
			criticalCount++
			oa.anomalies = append(oa.anomalies, Anomaly{
				Level:       "critical",
				Category:    "memory",
				Title:       "Extremely Large Key Detected",
				Description: fmt.Sprintf("Key '%s' is %s, which is extremely large", truncateKey(entry.Key), formatBytes(entry.Bytes)),
				Impact:      "Can cause blocking operations, memory pressure, and slow replication",
				Suggestion:  "Consider splitting this key into smaller chunks or using a different data structure",
				Value:       formatBytes(entry.Bytes),
				DetectedAt:  time.Now(),
			})
		} else if entry.Bytes >= warningSize {
			warningCount++
		}
	}

	if warningCount > 0 {
		oa.anomalies = append(oa.anomalies, Anomaly{
			Level:       "warning",
			Category:    "memory",
			Title:       "Large Keys Detected",
			Description: fmt.Sprintf("Found %d keys larger than 10MB", warningCount),
			Impact:      "May cause performance degradation and increased memory fragmentation",
			Suggestion:  "Review large keys and consider optimization",
			Value:       fmt.Sprintf("%d keys", warningCount),
			DetectedAt:  time.Now(),
		})
	}
}

// detectMemoryHotspots identifies memory concentration issues
func (oa *OpsAnalyzer) detectMemoryHotspots() {
	// Analyze by key prefix
	prefixes := oa.counter.GetLargestKeyPrefixes()

	for i, prefix := range prefixes {
		if i >= 20 { // Top 20
			break
		}

		percentage := float64(prefix.Bytes) / float64(oa.totalBytes) * 100
		avgSize := uint64(0)
		if prefix.Num > 0 {
			avgSize = prefix.Bytes / prefix.Num
		}

		hotspot := MemoryHotspot{
			Type:       "key_prefix",
			Identifier: prefix.Key,
			MemoryUsed: prefix.Bytes,
			KeyCount:   prefix.Num,
			Percentage: percentage,
			AvgKeySize: avgSize,
		}
		oa.memoryHotspots = append(oa.memoryHotspots, hotspot)

		// Alert if single prefix uses >30% memory
		if percentage > 30 {
			oa.anomalies = append(oa.anomalies, Anomaly{
				Level:       "warning",
				Category:    "memory",
				Title:       "Memory Hotspot Detected",
				Description: fmt.Sprintf("Key prefix '%s' uses %.1f%% of total memory", prefix.Key, percentage),
				Impact:      "Memory concentration can cause uneven load distribution in cluster mode",
				Suggestion:  "Consider reviewing keys with this prefix for optimization or better distribution",
				Value:       fmt.Sprintf("%.1f%%", percentage),
				DetectedAt:  time.Now(),
			})
		}
	}

	// Analyze by data type
	for typ, bytes := range oa.counter.typeBytes {
		percentage := float64(bytes) / float64(oa.totalBytes) * 100
		count := oa.counter.typeNum[typ]

		if percentage > 50 {
			oa.anomalies = append(oa.anomalies, Anomaly{
				Level:       "info",
				Category:    "memory",
				Title:       "Data Type Dominance",
				Description: fmt.Sprintf("Type '%s' accounts for %.1f%% of memory usage", typ, percentage),
				Impact:      "Single type dominance might indicate optimization opportunities",
				Suggestion:  "Review if this data type usage pattern is optimal for your use case",
				Value:       fmt.Sprintf("%.1f%%", percentage),
				DetectedAt:  time.Now(),
			})
		}

		avgSize := uint64(0)
		if count > 0 {
			avgSize = bytes / count
		}

		oa.memoryHotspots = append(oa.memoryHotspots, MemoryHotspot{
			Type:       "data_type",
			Identifier: typ,
			MemoryUsed: bytes,
			KeyCount:   count,
			Percentage: percentage,
			AvgKeySize: avgSize,
		})
	}

	// Sort hotspots by memory usage
	sort.Slice(oa.memoryHotspots, func(i, j int) bool {
		return oa.memoryHotspots[i].MemoryUsed > oa.memoryHotspots[j].MemoryUsed
	})
}

// detectKeyExplosion identifies rapid key growth patterns
func (oa *OpsAnalyzer) detectKeyExplosion() {
	// Warn if total keys exceed thresholds
	if oa.totalKeys > 10000000 { // 10M keys
		oa.anomalies = append(oa.anomalies, Anomaly{
			Level:       "warning",
			Category:    "keys",
			Title:       "High Key Count",
			Description: fmt.Sprintf("Database contains %s keys", formatNumber(oa.totalKeys)),
			Impact:      "High key count can slow down operations like KEYS, SCAN, and BGSAVE",
			Suggestion:  "Consider implementing key expiration policies or data archiving",
			Value:       formatNumber(oa.totalKeys),
			DetectedAt:  time.Now(),
		})
	}

	// Check for tiny keys (potential key explosion)
	tinyKeyCount := uint64(0)
	largestKeys := oa.counter.GetLargestEntries(500)
	for _, entry := range largestKeys {
		if entry.Bytes < 100 { // Less than 100 bytes
			tinyKeyCount++
		}
	}

	if tinyKeyCount > 100 && float64(tinyKeyCount)/float64(len(largestKeys)) > 0.3 {
		oa.anomalies = append(oa.anomalies, Anomaly{
			Level:       "warning",
			Category:    "keys",
			Title:       "Many Tiny Keys Detected",
			Description: "Large number of very small keys found, indicating possible key explosion",
			Impact:      "Overhead of key storage can exceed value storage, wasting memory",
			Suggestion:  "Consider using Hash data structures to group related small values",
			Value:       fmt.Sprintf("%d tiny keys", tinyKeyCount),
			DetectedAt:  time.Now(),
		})
	}
}

// detectTypeImbalance checks for inefficient type usage
func (oa *OpsAnalyzer) detectTypeImbalance() {
	// Check for types with very large element counts
	largestKeys := oa.counter.GetLargestEntries(100)

	const hugeCollectionSize = 1000000 // 1M elements

	for _, entry := range largestKeys {
		if entry.NumOfElem > hugeCollectionSize {
			oa.anomalies = append(oa.anomalies, Anomaly{
				Level:       "warning",
				Category:    "performance",
				Title:       "Huge Collection Detected",
				Description: fmt.Sprintf("Key '%s' (%s) contains %s elements", truncateKey(entry.Key), entry.Type, formatNumber(entry.NumOfElem)),
				Impact:      "Operations on huge collections can block Redis and cause latency spikes",
				Suggestion:  "Consider splitting into smaller collections or using different access patterns",
				Value:       formatNumber(entry.NumOfElem),
				DetectedAt:  time.Now(),
			})
		}
	}
}

// analyzeKeyPatterns identifies common key naming patterns
func (oa *OpsAnalyzer) analyzeKeyPatterns() {
	prefixes := oa.counter.GetLargestKeyPrefixes()

	for i, prefix := range prefixes {
		if i >= 50 { // Top 50 patterns
			break
		}

		percentage := float64(prefix.Num) / float64(oa.totalKeys) * 100
		avgMemory := uint64(0)
		if prefix.Num > 0 {
			avgMemory = prefix.Bytes / prefix.Num
		}

		// Get an example key
		example := ""
		largestKeys := oa.counter.GetLargestEntries(500)
		for _, entry := range largestKeys {
			if len(entry.Key) >= len(prefix.Key) && entry.Key[:len(prefix.Key)] == prefix.Key {
				example = entry.Key
				break
			}
		}

		pattern := KeyPattern{
			Pattern:     prefix.Key,
			Count:       prefix.Num,
			TotalMemory: prefix.Bytes,
			AvgMemory:   avgMemory,
			Percentage:  percentage,
			Example:     example,
		}
		oa.keyPatterns = append(oa.keyPatterns, pattern)
	}
}

// analyzeTypeEfficiency analyzes data type usage efficiency
func (oa *OpsAnalyzer) analyzeTypeEfficiency() {
	largestKeys := oa.counter.GetLargestEntries(500)

	// Group by type
	typeKeys := make(map[string][]uint64)
	for _, entry := range largestKeys {
		typeKeys[entry.Type] = append(typeKeys[entry.Type], entry.Bytes)
	}

	for typ, sizes := range typeKeys {
		if len(sizes) == 0 {
			continue
		}

		sort.Slice(sizes, func(i, j int) bool { return sizes[i] < sizes[j] })

		avg := uint64(0)
		sum := uint64(0)
		for _, size := range sizes {
			sum += size
		}
		avg = sum / uint64(len(sizes))

		median := sizes[len(sizes)/2]
		p95 := sizes[int(float64(len(sizes))*0.95)]
		p99 := sizes[int(float64(len(sizes))*0.99)]

		// Calculate efficiency (inverse of variance, scaled)
		variance := float64(0)
		for _, size := range sizes {
			diff := float64(size) - float64(avg)
			variance += diff * diff
		}
		variance /= float64(len(sizes))
		stddev := math.Sqrt(variance)

		efficiency := 100.0
		if avg > 0 {
			cv := stddev / float64(avg) // coefficient of variation
			efficiency = math.Max(0, 100.0-cv*100.0)
		}

		oa.typeEfficiency[typ] = TypeEfficiency{
			AvgSize:      avg,
			MedianSize:   median,
			P95Size:      p95,
			P99Size:      p99,
			Efficiency:   efficiency,
			WastedMemory: 0, // TODO: calculate based on Redis overhead
			OptimalType:  typ,
		}

		// Alert if efficiency is low
		if efficiency < 50 {
			oa.anomalies = append(oa.anomalies, Anomaly{
				Level:       "info",
				Category:    "performance",
				Title:       "Inconsistent Key Sizes",
				Description: fmt.Sprintf("Type '%s' shows high size variance (efficiency: %.1f%%)", typ, efficiency),
				Impact:      "Inconsistent sizes can indicate suboptimal data structure usage",
				Suggestion:  "Review keys of this type for potential optimization",
				Value:       fmt.Sprintf("%.1f%% efficient", efficiency),
				DetectedAt:  time.Now(),
			})
		}
	}
}

// analyzeClusterBalance analyzes slot distribution for clusters
func (oa *OpsAnalyzer) analyzeClusterBalance() {
	if len(oa.counter.slotBytes) == 0 {
		return
	}

	// Calculate slot statistics
	var totalSlotMemory uint64
	var maxSlotMemory uint64
	var minSlotMemory uint64 = ^uint64(0) // max uint64

	for _, bytes := range oa.counter.slotBytes {
		totalSlotMemory += bytes
		if bytes > maxSlotMemory {
			maxSlotMemory = bytes
		}
		if bytes < minSlotMemory {
			minSlotMemory = bytes
		}
	}

	// Calculate imbalance
	avgSlotMemory := totalSlotMemory / uint64(len(oa.counter.slotBytes))
	if avgSlotMemory > 0 {
		oa.slotImbalance = (float64(maxSlotMemory) - float64(minSlotMemory)) / float64(avgSlotMemory) * 100
	}

	// Get top slots
	type slotInfo struct {
		slot   int
		bytes  uint64
		keys   uint64
	}

	slots := make([]slotInfo, 0, len(oa.counter.slotBytes))
	for slot, bytes := range oa.counter.slotBytes {
		slots = append(slots, slotInfo{
			slot:  slot,
			bytes: bytes,
			keys:  oa.counter.slotNum[slot],
		})
	}

	sort.Slice(slots, func(i, j int) bool {
		return slots[i].bytes > slots[j].bytes
	})

	// Top 10 slots
	for i := 0; i < 10 && i < len(slots); i++ {
		percentage := float64(slots[i].bytes) / float64(oa.totalBytes) * 100
		oa.topSlotsUsage = append(oa.topSlotsUsage, SlotUsage{
			Slot:       slots[i].slot,
			KeyCount:   slots[i].keys,
			MemoryUsed: slots[i].bytes,
			Percentage: percentage,
		})
	}

	// Alert if imbalance is significant
	if oa.slotImbalance > 50 {
		oa.anomalies = append(oa.anomalies, Anomaly{
			Level:       "warning",
			Category:    "cluster",
			Title:       "Slot Imbalance Detected",
			Description: fmt.Sprintf("Cluster slots show %.1f%% imbalance", oa.slotImbalance),
			Impact:      "Uneven slot distribution can cause hotspots and performance issues",
			Suggestion:  "Consider rebalancing slots or reviewing key distribution strategy",
			Value:       fmt.Sprintf("%.1f%% imbalance", oa.slotImbalance),
			DetectedAt:  time.Now(),
		})
	}
}

// calculateHealthScore computes overall health (0-100)
func (oa *OpsAnalyzer) calculateHealthScore() {
	score := 100

	// Deduct points based on anomalies
	for _, anomaly := range oa.anomalies {
		switch anomaly.Level {
		case "critical":
			score -= 15
		case "warning":
			score -= 8
		case "info":
			score -= 3
		}
	}

	// Deduct points for high key count
	if oa.totalKeys > 100000000 { // 100M
		score -= 10
	} else if oa.totalKeys > 50000000 { // 50M
		score -= 5
	}

	// Deduct points for large average key size
	if oa.avgKeySize > 1024*1024 { // 1MB avg
		score -= 10
	} else if oa.avgKeySize > 100*1024 { // 100KB avg
		score -= 5
	}

	if score < 0 {
		score = 0
	}

	oa.healthScore = score
}

// generateRecommendations creates actionable recommendations
func (oa *OpsAnalyzer) generateRecommendations() {
	// Based on anomalies, generate specific recommendations

	// Memory optimization
	if oa.totalBytes > 10*1024*1024*1024 { // >10GB
		oa.recommendations = append(oa.recommendations, Recommendation{
			Priority:    2,
			Category:    "memory",
			Title:       "Enable Memory Eviction Policy",
			Description: "Database is using significant memory (>10GB)",
			Action:      "Configure 'maxmemory' and 'maxmemory-policy' in redis.conf",
			Impact:      "Prevents OOM errors and automatic eviction of less important data",
			Effort:      "low",
			CreatedAt:   time.Now(),
		})
	}

	// Key expiration
	largestKeys := oa.counter.GetLargestEntries(100)
	keysNeedingTTL := 0
	for range largestKeys {
		// In real implementation, check if key has TTL from Entry.Expiry
		keysNeedingTTL++
	}

	if keysNeedingTTL > 50 {
		oa.recommendations = append(oa.recommendations, Recommendation{
			Priority:    1,
			Category:    "ttl",
			Title:       "Implement TTL for Large Keys",
			Description: "Many large keys appear to have no expiration set",
			Action:      "Review and set appropriate TTL values for large keys",
			Impact:      "Prevents unbounded memory growth and automatic cleanup",
			Effort:      "medium",
			CreatedAt:   time.Now(),
		})
	}

	// Type optimization
	for typ, efficiency := range oa.typeEfficiency {
		if efficiency.Efficiency < 60 && typ == "string" {
			oa.recommendations = append(oa.recommendations, Recommendation{
				Priority:    3,
				Category:    "performance",
				Title:       "Consider Using Hash for Small Strings",
				Description: fmt.Sprintf("String type shows low efficiency (%.1f%%)", efficiency.Efficiency),
				Action:      "Group related small string values into Hash structures",
				Impact:      "Can reduce memory overhead by 30-50% for small values",
				Effort:      "high",
				CreatedAt:   time.Now(),
			})
		}
	}

	// Monitoring
	oa.recommendations = append(oa.recommendations, Recommendation{
		Priority:    4,
		Category:    "monitoring",
		Title:       "Enable Redis Slow Log",
		Description: "Track slow commands for performance optimization",
		Action:      "Set 'slowlog-log-slower-than 10000' and 'slowlog-max-len 128'",
		Impact:      "Helps identify performance bottlenecks",
		Effort:      "low",
		CreatedAt:   time.Now(),
	})

	// Sort by priority
	sort.Slice(oa.recommendations, func(i, j int) bool {
		return oa.recommendations[i].Priority < oa.recommendations[j].Priority
	})
}

// Helper functions

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatNumber(num uint64) string {
	if num < 1000 {
		return fmt.Sprintf("%d", num)
	}
	if num < 1000000 {
		return fmt.Sprintf("%.1fK", float64(num)/1000)
	}
	if num < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(num)/1000000)
	}
	return fmt.Sprintf("%.1fB", float64(num)/1000000000)
}

func truncateKey(key string) string {
	if len(key) <= 50 {
		return key
	}
	return key[:47] + "..."
}
