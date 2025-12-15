package dump

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

// HistoryEntry represents a single analysis history entry
type HistoryEntry struct {
	Filename    string    `json:"filename"`
	FilePath    string    `json:"filepath"`
	UploadTime  time.Time `json:"upload_time"`
	FileSize    int64     `json:"file_size"`
	TotalKeys   uint64    `json:"total_keys"`
	TotalMemory uint64    `json:"total_memory"`
}

// HistoryManager manages analysis history
type HistoryManager struct {
	entries  []HistoryEntry
	mu       sync.RWMutex
	filePath string
}

var historyManager *HistoryManager

// InitHistoryManager initializes the history manager
func InitHistoryManager(historyFile string) {
	historyManager = &HistoryManager{
		entries:  []HistoryEntry{},
		filePath: historyFile,
	}
	historyManager.load()
}

// GetHistoryManager returns the global history manager
func GetHistoryManager() *HistoryManager {
	if historyManager == nil {
		InitHistoryManager("history.json")
	}
	return historyManager
}

// Add adds a new entry to history
func (hm *HistoryManager) Add(entry HistoryEntry) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// Check if entry already exists (by filename)
	for i, e := range hm.entries {
		if e.Filename == entry.Filename {
			// Update existing entry
			hm.entries[i] = entry
			return hm.save()
		}
	}

	// Add new entry at the beginning (most recent first)
	hm.entries = append([]HistoryEntry{entry}, hm.entries...)

	// Keep only last 100 entries
	if len(hm.entries) > 100 {
		hm.entries = hm.entries[:100]
	}

	return hm.save()
}

// Get returns an entry by filename
func (hm *HistoryManager) Get(filename string) (HistoryEntry, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	for _, e := range hm.entries {
		if e.Filename == filename {
			return e, true
		}
	}
	return HistoryEntry{}, false
}

// GetAll returns all history entries
func (hm *HistoryManager) GetAll() []HistoryEntry {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Return a copy to prevent external modifications
	result := make([]HistoryEntry, len(hm.entries))
	copy(result, hm.entries)
	return result
}

// Remove removes an entry by filename
func (hm *HistoryManager) Remove(filename string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	for i, e := range hm.entries {
		if e.Filename == filename {
			hm.entries = append(hm.entries[:i], hm.entries[i+1:]...)
			return hm.save()
		}
	}
	return nil
}

// load loads history from file
func (hm *HistoryManager) load() error {
	data, err := ioutil.ReadFile(hm.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, that's ok
			return nil
		}
		log.Printf("Error loading history file: %v", err)
		return err
	}

	if len(data) == 0 {
		return nil
	}

	err = json.Unmarshal(data, &hm.entries)
	if err != nil {
		log.Printf("Error parsing history file: %v", err)
		return err
	}

	log.Printf("Loaded %d history entries", len(hm.entries))
	return nil
}

// save saves history to file
func (hm *HistoryManager) save() error {
	data, err := json.MarshalIndent(hm.entries, "", "  ")
	if err != nil {
		log.Printf("Error marshaling history: %v", err)
		return err
	}

	err = ioutil.WriteFile(hm.filePath, data, 0644)
	if err != nil {
		log.Printf("Error saving history file: %v", err)
		return err
	}

	return nil
}
