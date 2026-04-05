package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const serverURL = "http://localhost:8080"

func main() {
	id := flag.String("id", "", "The ID for the CIM model operations")
	flag.Parse()

	if *id == "" {
		fmt.Println("Error: -id flag is required")
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Error: a command (upload, get) is required")
		os.Exit(1)
	}

	command := args[0]
	switch command {
	case "upload":
		if len(args) < 2 {
			fmt.Println("Error: upload command requires at least one file path")
			os.Exit(1)
		}
		uploadFiles(*id, args[1:])
	case "get":
		getData(*id)
	default:
		fmt.Printf("Error: unknown command '%s'\n", command)
		os.Exit(1)
	}
}

func uploadFiles(id string, filePaths []string) {
	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Error opening file %s: %v\n", filePath, err)
			continue
		}
		defer file.Close()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			fmt.Printf("Error creating form file for %s: %v\n", filePath, err)
			continue
		}

		if _, err = io.Copy(part, file); err != nil {
			fmt.Printf("Error copying file content for %s: %v\n", filePath, err)
			continue
		}
		writer.Close()

		url := fmt.Sprintf("%s/cim/%s", serverURL, id)
		req, err := http.NewRequest("POST", url, body)
		if err != nil {
			fmt.Printf("Error creating request for %s: %v\n", filePath, err)
			continue
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error uploading file %s: %v\n", filePath, err)
			continue
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error response from server for %s (HTTP %d): %s\n", filePath, resp.StatusCode, string(respBody))
		} else {
			fmt.Printf("Successfully uploaded %s\n", filePath)
		}
	}
}

func getData(id string) {
	url := fmt.Sprintf("%s/cim/%s", serverURL, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating get request: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending get request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error response from server (HTTP %d): %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err != nil {
		// If indenting fails, just copy the raw body
		fmt.Println(string(bodyBytes))
		return
	}

	fmt.Println(prettyJSON.String())
}
