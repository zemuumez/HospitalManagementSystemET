package postgres

import (
	"context"
	"errors"
	"testing"

	"hms.local/api/internal/domain"
)

func testOverviewScope(t *testing.T, store Store, actors []domain.Actor) {
	t.Helper()
	ctx := context.Background()
	admin, err := store.Overview(ctx, actors[0])
	if err != nil || admin.InvoicesMinor <= 0 || admin.Doctors == 0 {
		t.Fatalf("administrator overview must retain populated totals: %+v, %v", admin, err)
	}
	for _, actor := range actors[1:] {
		t.Run("overview scope "+actor.Role+" "+actor.ID, func(t *testing.T) {
			got, err := store.Overview(ctx, actor)
			if err != nil {
				t.Fatal(err)
			}
			var count, today int
			if err := store.DB.QueryRow(ctx, `SELECT count(*), count(*) FILTER(WHERE created_at::date=CURRENT_DATE) FROM patient WHERE canonical_patient_id(id)=id AND `+scope, actor.Role, actor.ID).Scan(&count, &today); err != nil {
				t.Fatal(err)
			}
			want := domain.Overview{PatientCount: count, RegisteredToday: today, Patients: count}
			if got != want {
				t.Fatalf("overview exposes values outside patient scope: got %+v want %+v", got, want)
			}
		})
	}
	for _, role := range []string{"accountant", "nurse", "pharmacist", "lab_technician", "case_manager", "unknown"} {
		if _, err := store.Overview(ctx, domain.Actor{Role: role}); !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("unauthorized overview for %s: %v", role, err)
		}
	}
	var available int
	if err := store.DB.QueryRow(ctx, `SELECT count(*) FROM hospital_bed b WHERE b.active AND b.state='ready' AND NOT EXISTS(SELECT 1 FROM encounter e WHERE e.bed_id=b.id AND e.status='active')`).Scan(&available); err != nil {
		t.Fatal(err)
	}
	if admin.AvailableBeds != available {
		t.Fatalf("available beds includes non-ready beds: got %d want %d", admin.AvailableBeds, available)
	}
}
