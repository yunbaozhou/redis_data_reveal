package dump

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// opsAnalysisHandler returns comprehensive operational analysis
func opsAnalysisHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	path := p.ByName("path")
	c := counters.Get(path)
	if c == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Instance not found or still parsing",
		})
		return
	}

	counter := c.(*Counter)
	analyzer := NewOpsAnalyzer(counter)

	response := map[string]interface{}{
		"health_score":         analyzer.healthScore,
		"anomalies":            analyzer.anomalies,
		"memory_hotspots":      analyzer.memoryHotspots,
		"key_patterns":         analyzer.keyPatterns,
		"type_efficiency":      analyzer.typeEfficiency,
		"slot_imbalance":       analyzer.slotImbalance,
		"top_slots_usage":      analyzer.topSlotsUsage,
		"recommendations":      analyzer.recommendations,
		"basic_stats": map[string]interface{}{
			"total_keys":     analyzer.totalKeys,
			"total_bytes":    analyzer.totalBytes,
			"avg_key_size":   analyzer.avgKeySize,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// opsAnomaliesHandler returns only anomalies for quick alerts
func opsAnomaliesHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	path := p.ByName("path")
	c := counters.Get(path)
	if c == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Instance not found",
		})
		return
	}

	counter := c.(*Counter)
	analyzer := NewOpsAnalyzer(counter)

	// Group anomalies by level
	critical := []Anomaly{}
	warnings := []Anomaly{}
	info := []Anomaly{}

	for _, anomaly := range analyzer.anomalies {
		switch anomaly.Level {
		case "critical":
			critical = append(critical, anomaly)
		case "warning":
			warnings = append(warnings, anomaly)
		case "info":
			info = append(info, anomaly)
		}
	}

	response := map[string]interface{}{
		"critical": critical,
		"warning":  warnings,
		"info":     info,
		"total":    len(analyzer.anomalies),
		"health_score": analyzer.healthScore,
	}

	json.NewEncoder(w).Encode(response)
}

// opsRecommendationsHandler returns only recommendations
func opsRecommendationsHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	path := p.ByName("path")
	c := counters.Get(path)
	if c == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Instance not found",
		})
		return
	}

	counter := c.(*Counter)
	analyzer := NewOpsAnalyzer(counter)

	response := map[string]interface{}{
		"recommendations": analyzer.recommendations,
		"total":          len(analyzer.recommendations),
	}

	json.NewEncoder(w).Encode(response)
}

// opsHealthHandler returns quick health check
func opsHealthHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")

	path := p.ByName("path")
	c := counters.Get(path)
	if c == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Instance not found",
		})
		return
	}

	counter := c.(*Counter)
	analyzer := NewOpsAnalyzer(counter)

	healthStatus := "excellent"
	if analyzer.healthScore < 90 {
		healthStatus = "good"
	}
	if analyzer.healthScore < 75 {
		healthStatus = "fair"
	}
	if analyzer.healthScore < 60 {
		healthStatus = "poor"
	}
	if analyzer.healthScore < 40 {
		healthStatus = "critical"
	}

	criticalCount := 0
	warningCount := 0
	for _, anomaly := range analyzer.anomalies {
		if anomaly.Level == "critical" {
			criticalCount++
		} else if anomaly.Level == "warning" {
			warningCount++
		}
	}

	response := map[string]interface{}{
		"health_score":     analyzer.healthScore,
		"health_status":    healthStatus,
		"critical_issues":  criticalCount,
		"warnings":         warningCount,
		"total_anomalies":  len(analyzer.anomalies),
		"recommendations":  len(analyzer.recommendations),
	}

	json.NewEncoder(w).Encode(response)
}
