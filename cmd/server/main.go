package main

import (
	"archive/zip"
	"bytes"
	"cimgo/cimstructs"
	"cimgo/cgmesxml"
	"cimgo/cimproto"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
)

// Global map to store CIM specifications in memory, keyed by ID.
var cimDataset = make(map[string]*cimstructs.CIMElementList)
var protoDataset = make(map[string][]byte)
var mu sync.RWMutex

const uploadDir = "/tmp/cimgo-uploads"

func main() {
	http.HandleFunc("/cim/", handleCIMRequest)
	log.Println("Starting webserver on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleCIMRequest(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/cim/"), "/")
	id := parts[0]
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPost:
		uploadFileHandler(w, r, id)
	case http.MethodGet:
		if len(parts) > 1 && parts[1] == "proto" {
			getProtoHandler(w, r, id)
		} else {
			getCIMHandler(w, r, id)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func uploadFileHandler(w http.ResponseWriter, r *http.Request, id string) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB
		http.Error(w, "Error parsing multipart form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	dirPath := filepath.Join(uploadDir, id)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(dirPath, handler.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	if strings.HasSuffix(handler.Filename, ".zip") {
		if err := unzip(filePath, dirPath); err != nil {
			http.Error(w, fmt.Sprintf("Failed to unzip archive: %v", err), http.StatusInternalServerError)
			return
		}
		// Optionally remove the zip file after extraction
		// os.Remove(filePath)
	}
	log.Printf("File %s uploaded successfully for ID %s\n", handler.Filename, id)

	// Trigger processing asynchronously
	go func() {
		if err := processCIMFiles(id); err != nil {
			log.Printf("Error processing files for ID %s: %v", id, err)
		}
	}()

	fmt.Fprintf(w, "File %s uploaded successfully for ID %s\n", handler.Filename, id)
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func processCIMFiles(id string) error {
	dirPath := filepath.Join(uploadDir, id)
	log.Printf("Processing files in directory: %s\n", dirPath)
	globPattern := filepath.Join(dirPath, "*.xml") // Assuming XML files

	entries, err := filepath.Glob(globPattern)
	if err != nil {
		return fmt.Errorf("glob failed: %w", err)
	}
	log.Println("Processing files:", entries)

	readers := make([]io.Reader, 0, len(entries))
	for _, entry := range entries {
		b, err := os.ReadFile(entry)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", entry, err)
		}
		readers = append(readers, bytes.NewReader(b))
	}

	mergedCIMData, err := cgmesxml.DecodeProfiles(readers, nil)
	if err != nil {
		return fmt.Errorf("failed to decode profiles: %w", err)
	}
	log.Println("Decoded CIM data")

	jsonOut, err := json.MarshalIndent(mergedCIMData.Elements, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to create a nicely formatted JSON: %w", err)
	}
	debugFile := filepath.Join(dirPath, "merged_output.json")
	if err := os.WriteFile(debugFile, jsonOut, 0644); err != nil {
		log.Printf("Failed to write merged output to disk: %v", err)
	}

	// Generate Protobuf serialization
	protoList, err := cimproto.ToProto(mergedCIMData)
	if err != nil {
		return fmt.Errorf("failed to convert to proto: %w", err)
	}

	protoData, err := proto.Marshal(protoList)
	if err != nil {
		return fmt.Errorf("failed to marshal to protobuf: %w", err)
	}

	// Write the proto data back to disk for debugging
	protoFile := filepath.Join(dirPath, "merged_output.bin")
	if err := os.WriteFile(protoFile, protoData, 0644); err != nil {
		log.Printf("Failed to write merged proto output to disk: %v", err)
	}

	mu.Lock()
	cimDataset[id] = mergedCIMData
	protoDataset[id] = protoData
	mu.Unlock()

	return nil
}

func getCIMHandler(w http.ResponseWriter, r *http.Request, id string) {
	mu.RLock()
	cimData, ok := cimDataset[id]
	mu.RUnlock()

	if !ok {
		http.Error(w, "No data found for this ID", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cimData); err != nil {
		http.Error(w, "Failed to serialize data", http.StatusInternalServerError)
		return
	}
}

func getProtoHandler(w http.ResponseWriter, r *http.Request, id string) {
	mu.RLock()
	protoData, ok := protoDataset[id]
	mu.RUnlock()

	if !ok {
		http.Error(w, "No data found for this ID", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.bin", id))
	w.Write(protoData)
}
