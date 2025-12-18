package integration

import (
	"sync"
	"testing"
	"time"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/repository"
	"github.com/Fixsbreaker/event-hub/backend/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestRegistrationRaceCondition verifies that concurrent registrations respect capacity
func TestRegistrationRaceCondition(t *testing.T) {
	// 1. Setup DB (InMemory SQLite support transactions?)
	// SQLite supports transactions, but FOR UPDATE might be ignored or behave differently than Postgres.
	// However, GORM transaction mutex should still hold within the same process if configured correctly,
	// OR we might need to rely on the fact that we use "lock" logic.
	// Real test requires Postgres, but let's try with SQLite first to check logic flow.

	// Connect to Docker Postgres (port 5430 as per our previous verification)
	dsn := "host=localhost user=postgres password=icantbelieveyoureadthis dbname=event_hub port=5430 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	// Skip AutoMigrate as DB is already set up.
	// Since we use random UUIDs for every run, collisions are unlikely, so no need to explicit delete.

	// Repos & Services
	regRepo := repository.NewRegistrationRepository(db)
	eventRepo := repository.NewEventRepository(db)
	regService := service.NewRegistrationService(regRepo, eventRepo)

	// Generate valid UUIDs
	eventID := uuid.NewString()
	organizerID := uuid.NewString()

	// Create Organizer User to satisfy FK
	err = db.Exec("INSERT INTO users (id, email, password_hash, name, role) VALUES (?, ?, ?, ?, ?)",
		organizerID, "org-"+organizerID+"@test.com", "hash", "Organizer", "organizer").Error
	require.NoError(t, err)

	err = eventRepo.Create(&domain.Event{
		ID:            eventID,
		OrganizerID:   organizerID,
		Title:         "Race Event",
		Capacity:      1,
		Status:        "published",
		StartDatetime: time.Now().Add(1 * time.Hour),
		EndDatetime:   time.Now().Add(2 * time.Hour),
	})
	require.NoError(t, err)

	// 3. Launch 2 concurrent goroutines trying to register
	var wg sync.WaitGroup
	results := make(chan error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		// Generate valid User UUID
		userID := uuid.NewString()

		// Create Registrant User to satisfy FK
		err := db.Exec("INSERT INTO users (id, email, password_hash, name, role) VALUES (?, ?, ?, ?, ?)",
			userID, "user-"+userID+"@test.com", "hash", "User", "user").Error
		require.NoError(t, err)

		go func(uid string) {
			defer wg.Done()
			_, err := regService.RegisterUser(uid, eventID)
			results <- err
		}(userID)
	}

	wg.Wait()
	close(results)

	// 4. Verify results
	successCount := 0
	errorCount := 0
	for err := range results {
		if err == nil {
			successCount++
		} else {
			errorCount++
		}
	}

	// We expect exactly 1 success and 1 error (Capacity 1)
	require.Equal(t, 1, successCount, "Expected exactly 1 successful registration")
	require.Equal(t, 1, errorCount, "Expected exactly 1 failed registration")

	// Double check count
	count, _ := regRepo.CountByEvent(eventID)
	require.Equal(t, int64(1), count, "DB should have exactly 1 registration")
}
