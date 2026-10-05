package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type FollowUp struct {
	Course    string    `json:"course"`
	Learner   string    `json:"learner"`
	Deadline  time.Time `json:"deadline"`
	Educator  string    `json:"educator"`
	DelayHour int       `json:"delay_hours"`
}

func (f FollowUp) dueAt() time.Time { return f.Deadline.Add(time.Duration(f.DelayHour) * time.Hour) }

type Service struct {
	client   *InfraiClient
	mu       sync.Mutex
	accepted int
}

func (s *Service) schedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method must be POST", http.StatusMethodNotAllowed)
		return
	}
	var f FollowUp
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil || f.Course == "" || f.Learner == "" || f.Educator == "" || f.DelayHour < 0 {
		http.Error(w, "course, learner, educator and delay_hours are required", http.StatusBadRequest)
		return
	}
	payload := map[string]any{"course": f.Course, "learner": f.Learner, "educator": f.Educator, "send_at": f.dueAt().UTC().Format(time.RFC3339)}
	if err := s.client.Publish(payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	s.mu.Lock()
	s.accepted++
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"scheduled","send_at":%q}`, f.dueAt().UTC().Format(time.RFC3339))
}

func (s *Service) report(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	count := s.accepted
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"scheduled_followups":%d}`, count)
}
