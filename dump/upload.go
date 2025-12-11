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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/dongmx/rdb"
	"github.com/julienschmidt/httprouter"
	"github.com/xueqiu/rdr/decoder"
)

const maxUploadSize = 10 * 1024 * 1024 * 1024 // 10GB

var uploadMutex sync.Mutex

type UploadResponse struct {
	Success   bool     `json:"success"`
	Message   string   `json:"message"`
	Instances []string `json:"instances"`
}

// uploadHandler handles file upload requests
func uploadHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Don't use MaxBytesReader for large files, let ParseMultipartForm handle it
	// Parse multipart form with large max memory (100MB in memory, rest on disk)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		respondWithError(w, "文件解析失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get uploaded files
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		respondWithError(w, "没有上传文件", http.StatusBadRequest)
		return
	}

	uploadMutex.Lock()
	defer uploadMutex.Unlock()

	// Create uploads directory if not exists
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Error creating upload directory: %v", err)
		respondWithError(w, "创建上传目录失败", http.StatusInternalServerError)
		return
	}

	var instances []string

	// Process each uploaded file
	for _, fileHeader := range files {
		// Validate file extension
		if filepath.Ext(fileHeader.Filename) != ".rdb" {
			log.Printf("Skipping non-rdb file: %v", fileHeader.Filename)
			continue
		}

		log.Printf("Processing upload: %v (size: %d bytes)", fileHeader.Filename, fileHeader.Size)

		// Open uploaded file
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Error opening uploaded file: %v", err)
			continue
		}

		// Create destination file
		destPath := filepath.Join(uploadDir, fileHeader.Filename)
		destFile, err := os.Create(destPath)
		if err != nil {
			log.Printf("Error creating destination file: %v", err)
			file.Close()
			continue
		}

		// Copy file content with buffer for large files
		written, err := io.Copy(destFile, file)
		file.Close()
		destFile.Close()

		if err != nil {
			log.Printf("Error copying file: %v", err)
			os.Remove(destPath)
			continue
		}

		log.Printf("File saved successfully: %v (%d bytes written)", fileHeader.Filename, written)

		// Parse the uploaded file
		filename := fileHeader.Filename
		if !counters.Check(filename) {
			instances = append(instances, filename)

			// Create progress tracker
			progress := NewParseProgress(filename)
			progress.AddLog(fmt.Sprintf("File uploaded: %s (%.2f MB)", filename, float64(written)/(1024*1024)))
			progress.SetStatus("parsing")

			// Start decoding in background
			go func(path, name string, pp *ParseProgress) {
				dec := decoder.NewDecoder()
				pp.AddLog("Initializing RDB decoder...")
				pp.SetProgress(5)

				// Start decoding in a goroutine
				go func() {
					// Note: rdb.Decode() will close dec.Entries internally, so we don't need to close it here
					pp.AddLog("Opening RDB file...")
					pp.SetProgress(10)

					f, err := os.Open(path)
					if err != nil {
						log.Printf("Error opening file %v: %v", name, err)
						pp.SetError(fmt.Sprintf("Failed to open file: %v", err))
						pp.AddLog(fmt.Sprintf("ERROR: %v", err))
						return
					}
					defer f.Close()

					pp.AddLog("Starting RDB decode process...")
					pp.SetProgress(20)

					err = rdb.Decode(f, dec)
					if err != nil {
						log.Printf("Error decoding file %v: %v", name, err)
						pp.SetError(fmt.Sprintf("Decode failed: %v", err))
						pp.AddLog(fmt.Sprintf("ERROR: %v", err))
						return
					}
					pp.AddLog("RDB decode completed successfully")
					pp.SetProgress(70)
				}()

				// Count entries (this will block until channel is closed)
				pp.AddLog("Counting and analyzing entries...")
				pp.SetProgress(30)

				counter := NewCounter()
				counter.Count(dec.Entries)

				pp.AddLog("Saving statistics...")
				pp.SetProgress(90)

				counters.Set(name, counter)
				log.Printf("Parse completed and counter saved: %v", name)

				pp.AddLog("Analysis complete!")
				pp.SetProgress(100)
				pp.SetStatus("completed")

				// Update template data
				if instances, ok := tplCommonData["Instances"].([]string); ok {
					tplCommonData["Instances"] = append(instances, name)
				} else {
					tplCommonData["Instances"] = []string{name}
				}
			}(destPath, filename, progress)
		} else {
			log.Printf("File already parsed: %v", filename)
			instances = append(instances, filename)
		}
	}

	if len(instances) == 0 {
		respondWithError(w, "没有有效的RDB文件", http.StatusBadRequest)
		return
	}

	// Send success response
	response := UploadResponse{
		Success:   true,
		Message:   "文件上传成功",
		Instances: instances,
	}

	json.NewEncoder(w).Encode(response)
}

// showUploadPage renders the upload page
func showUploadPage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile(w, r, "views/upload.html")
}

func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	response := UploadResponse{
		Success: false,
		Message: message,
	}
	json.NewEncoder(w).Encode(response)
}

// DecodeFile decodes a single RDB file
func DecodeFile(decoder *decoder.Decoder, filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("open %s failed, error: %v", filepath, err)
	}
	defer f.Close()

	// Import the rdb package
	if err := rdb.Decode(f, decoder); err != nil {
		return fmt.Errorf("decode failed: %v", err)
	}

	return nil
}
