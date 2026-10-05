package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"sync"
	"testing"
	"time"
)

func testQueueConcurrency(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor) {
	ctx := context.Background()
	app := application.Hospital{Store: store, Now: time.Now}
	ids := []string{}
	for n := 0; n < 16; n++ {
		p, e := app.Register(ctx, actors[0], domain.PatientInput{GivenName: "Queue", FamilyName: "Fixture", DateOfBirth: "2000-01-01"})
		if e != nil {
			t.Fatal(e)
		}
		ids = append(ids, p.ID)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	tokens := make(chan int, len(ids))
	for _, id := range ids {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			<-start
			item, e := store.EnqueuePatient(ctx, actors[0], domain.EnqueuePatientInput{DoctorID: actors[1].ID, PatientID: id, QueueDate: "2041-01-01"})
			if e != nil {
				t.Errorf("concurrent enqueue: %v", e)
				return
			}
			tokens <- item.TokenNumber
		}(id)
	}
	close(start)
	wg.Wait()
	close(tokens)
	seen := map[int]bool{}
	for token := range tokens {
		if seen[token] {
			t.Errorf("duplicate token %d", token)
		}
		seen[token] = true
	}
	if len(seen) != len(ids) {
		t.Fatalf("only %d/%d patients queued", len(seen), len(ids))
	}
}
