package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/ossrs/go-oryx-lib/errors"
)

func postImageBase64(ctx context.Context, url string, imageData string, v interface{}) error {
	// Create a JSON payload with the Base64 image
	payload := map[string]string{
		"content": imageData,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal JSON payload: %v", jsonPayload)
	}

	// Create the HTTP POST request
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonPayload))
	if err != nil {
		return errors.Wrapf(err, "api post failed, url=%v", url)
	}
	defer resp.Body.Close()
	
	// Check for non-200 status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return errors.Wrapf(err, "api returned non-200 status code, url=%v, status=%v", url, resp.Status)
	}

	// Read the response body
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err,"api read failed, url=%v", url)
	}

	if err := json.Unmarshal(b, v); err != nil {
		return errors.Wrapf(err, "json unmarshal %v", string(b))
	}

	return nil
}