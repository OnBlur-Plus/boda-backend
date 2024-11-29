package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func postImageBase64(url string, imagePath string) (responseBody []byte, err error) {
	// Open the image file
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %v", err)
	}
	defer file.Close()

	// Read the image file into memory
	imageData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %v", err)
	}

	// Convert the image data to Base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	// Create a JSON payload with the Base64 image
	payload := map[string]string{
		"image": base64Image,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON payload: %v", err)
	}

	// Create the HTTP POST request
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("api post failed, url=%v, err=%v", url, err)
	}
	defer resp.Body.Close()

	// Check for non-200 status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned non-200 status code, url=%v, status=%v", url, resp.Status)
	}

	// Read the response body
	responseBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("api read failed, url=%v, err=%v", url, err)
	}

	return responseBody, nil
}