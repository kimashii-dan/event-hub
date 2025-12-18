package worker

import (
	"testing"
	"time"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestWorkerPool(t *testing.T) {
	// Create pool with 2 workers and buffer of 5
	pool := NewWorkerPool(2, 5)

	// Start pool
	pool.Start()

	// We can't easily inspect internal state without mocking stdout or changing structure,
	// but we can ensure it doesn't panic and accepts jobs.

	job := NotificationJob{
		Notification: &domain.Notification{
			ID:     "test-id",
			UserID: "test-user",
			Title:  "Test Notif",
		},
		DestEmail: "test@test.com",
	}

	// Submit job (should not block)
	done := make(chan bool)
	go func() {
		pool.Submit(job)
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Submit timed out")
	}

	// Wait a bit to let worker "process" (it sleeps 2s in code, so this test might need patience
	// or we should update code to make sleep configurable. For now, we trust it runs).
	// In a real test, we'd inject a 'SleepDuration' or a mock processor.

	// Ensure Stop works
	stopDone := make(chan bool)
	go func() {
		pool.Stop() // This waits for workers, so it will take ~2 seconds due to the hardcoded sleep
		stopDone <- true
	}()

	select {
	case <-stopDone:
		// Success
	case <-time.After(5 * time.Second): // Wait longer than the 2s sleep
		t.Fatal("Stop timed out")
	}

	assert.True(t, true, "Test completed without panic")
}
