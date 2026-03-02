package uploader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HECEvent is the Splunk HEC payload.
type HECEvent struct {
	Time   float64     `json:"time"`
	Host   string      `json:"host"`
	Source string      `json:"source"`
	Index  string      `json:"index,omitempty"`
	Event  interface{} `json:"event"`
}

// HECUploader posts events to a Splunk HEC endpoint.
type HECUploader struct {
	url    string
	token  string
	client *http.Client
}

// New creates a new HECUploader.
func New(url, token string) *HECUploader {
	return &HECUploader{
		url:   url,
		token: token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send posts the event to Splunk HEC. Retries once after 20s on failure.
func (u *HECUploader) Send(evt HECEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal HEC event: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(20 * time.Second)
		}
		if err := u.post(data); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return lastErr
}

func (u *HECUploader) post(data []byte) error {
	req, err := http.NewRequest(http.MethodPost, u.url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Splunk "+u.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.client.Do(req)
	if err != nil {
		return fmt.Errorf("HEC POST: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HEC returned HTTP %d", resp.StatusCode)
	}
	return nil
}
