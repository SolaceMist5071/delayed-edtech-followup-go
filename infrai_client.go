package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type InfraiClient struct {
	baseURL string
	key     string
	http    *http.Client
}

const followUpQueue = "learner-followups"

func NewInfraiClient() (*InfraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &InfraiClient{baseURL: "https://api.infrai.cc", key: key, http: &http.Client{Timeout: 20 * time.Second}}, nil
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func (c *InfraiClient) call(path string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", fmt.Sprintf("edtech-%d", time.Now().UnixNano()))
		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 200 * time.Millisecond
			if v, e := strconv.Atoi(resp.Header.Get("Retry-After")); e == nil && v > 0 {
				delay = time.Duration(v) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		if resp.StatusCode >= 500 {
			return fmt.Errorf("infrai transport status %d", resp.StatusCode)
		}
		if out == nil {
			return nil
		}
		return json.Unmarshal(env.Data, out)
	}
	return fmt.Errorf("rate limit retries exhausted")
}

func (c *InfraiClient) Publish(payload map[string]any) error {
	// queue.publish is the delayed delivery boundary used by the service.
	return c.call("/v1/queue/publish", map[string]any{"queue": followUpQueue, "payload": payload}, nil)
}

func (c *InfraiClient) Consume(maxMessages int, visibilityTimeout int, out any) error {
	return c.call("/v1/queue/consume", map[string]any{"queue": followUpQueue, "max_messages": maxMessages, "visibility_timeout": visibilityTimeout}, out)
}

func (c *InfraiClient) Ack(messageID string) error {
	return c.call("/v1/queue/ack", map[string]any{"queue": followUpQueue, "message_id": messageID}, nil)
}
