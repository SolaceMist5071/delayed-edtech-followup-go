package main

import (
	"testing"
	"time"
)

func TestDueAtAddsDelayToLearnerDeadline(t *testing.T) {
	deadline := time.Date(2026, 9, 4, 9, 0, 0, 0, time.UTC)
	f := FollowUp{Course: "etl-101", Learner: "lee", Educator: "ana", Deadline: deadline, DelayHour: 6}
	want := time.Date(2026, 9, 4, 15, 0, 0, 0, time.UTC)
	if got := f.dueAt(); !got.Equal(want) {
		t.Fatalf("dueAt = %s, want %s", got, want)
	}
}
