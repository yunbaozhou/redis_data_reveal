// Copyright 2017 XUEQIU.COM
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dump

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
)

// ParseProgress tracks the parsing progress of a file
type ParseProgress struct {
	Filename    string
	Status      string // "pending", "parsing", "completed", "error"
	StartTime   time.Time
	Progress    int // 0-100
	CurrentStep string
	Logs        []string
	Error       string
	mu          sync.RWMutex
}

var (
	progressTrackers = make(map[string]*ParseProgress)
	progressMutex    sync.RWMutex
)

// NewParseProgress creates a new progress tracker
func NewParseProgress(filename string) *ParseProgress {
	progressMutex.Lock()
	defer progressMutex.Unlock()

	pp := &ParseProgress{
		Filename:  filename,
		Status:    "pending",
		StartTime: time.Now(),
		Progress:  0,
		Logs:      []string{},
	}
	progressTrackers[filename] = pp
	return pp
}

// GetProgress retrieves progress for a file
func GetProgress(filename string) *ParseProgress {
	progressMutex.RLock()
	defer progressMutex.RUnlock()
	return progressTrackers[filename]
}

// AddLog adds a log entry
func (pp *ParseProgress) AddLog(message string) {
	pp.mu.Lock()
	defer pp.mu.Unlock()
	timestamp := time.Now().Format("15:04:05")
	logEntry := fmt.Sprintf("[%s] %s", timestamp, message)
	pp.Logs = append(pp.Logs, logEntry)
}

// SetStatus updates the status
func (pp *ParseProgress) SetStatus(status string) {
	pp.mu.Lock()
	defer pp.mu.Unlock()
	pp.Status = status
}

// SetProgress updates the progress percentage
func (pp *ParseProgress) SetProgress(progress int) {
	pp.mu.Lock()
	defer pp.mu.Unlock()
	pp.Progress = progress
}

// SetCurrentStep updates the current step
func (pp *ParseProgress) SetCurrentStep(step string) {
	pp.mu.Lock()
	defer pp.mu.Unlock()
	pp.CurrentStep = step
}

// SetError sets an error message
func (pp *ParseProgress) SetError(err string) {
	pp.mu.Lock()
	defer pp.mu.Unlock()
	pp.Error = err
	pp.Status = "error"
}

// GetData returns progress data for JSON
func (pp *ParseProgress) GetData() map[string]interface{} {
	pp.mu.RLock()
	defer pp.mu.RUnlock()

	duration := time.Since(pp.StartTime)
	return map[string]interface{}{
		"filename":    pp.Filename,
		"status":      pp.Status,
		"progress":    pp.Progress,
		"currentStep": pp.CurrentStep,
		"logs":        pp.Logs,
		"error":       pp.Error,
		"duration":    duration.Seconds(),
	}
}

// progressHandler returns progress data as JSON
func progressHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	filename := p.ByName("path")

	pp := GetProgress(filename)
	if pp == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Progress not found"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	data := pp.GetData()

	// Convert to JSON manually for simplicity
	w.Write([]byte(fmt.Sprintf(`{
		"filename": "%s",
		"status": "%s",
		"progress": %d,
		"currentStep": "%s",
		"duration": %.1f,
		"error": "%s"
	}`, data["filename"], data["status"], data["progress"],
		data["currentStep"], data["duration"], data["error"])))
}

// streamLogsHandler streams logs using Server-Sent Events
func streamLogsHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	filename := p.ByName("path")

	pp := GetProgress(filename)
	if pp == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send initial logs
	pp.mu.RLock()
	lastLogIndex := len(pp.Logs)
	pp.mu.RUnlock()

	// Stream updates
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			pp.mu.RLock()
			currentLogCount := len(pp.Logs)
			status := pp.Status
			progress := pp.Progress

			// Send new logs
			if currentLogCount > lastLogIndex {
				for i := lastLogIndex; i < currentLogCount; i++ {
					fmt.Fprintf(w, "data: {\"type\":\"log\",\"message\":\"%s\"}\n\n", pp.Logs[i])
				}
				lastLogIndex = currentLogCount
				flusher.Flush()
			}

			// Send progress update
			fmt.Fprintf(w, "data: {\"type\":\"progress\",\"status\":\"%s\",\"progress\":%d}\n\n", status, progress)
			flusher.Flush()
			pp.mu.RUnlock()

			// Stop if completed or error
			if status == "completed" || status == "error" {
				time.Sleep(2 * time.Second)
				return
			}
		}
	}
}
