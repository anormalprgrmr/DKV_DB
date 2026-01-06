package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPResponse struct {
	StatusCode int
	Body       []byte
}

func GetObject(port int, key string) (*HTTPResponse, error) {
	url := fmt.Sprintf("http://localhost:%d/objects/%s", port, key)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
	}, nil
}

type PutRequestPayload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func PutObject(port int, key, value string) (*HTTPResponse, error) {
	url := fmt.Sprintf("http://localhost:%d/objects", port)

	payload := PutRequestPayload{
		Key:   key,
		Value: value,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(
		http.MethodPut,
		url,
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
	}, nil
}
