package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func testSchedulingChanges(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor, patients []domain.Patient) {
	t.Helper()
	ctx := context.Background()
	now := time.Now()
	s := application.Scheduling{Store: store, Now: func() time.Time { return now }}
	day := now.In(domain.HospitalLocation).AddDate(0, 0, 3)
	start := time.Date(day.Year(), day.Month(), day.Day(), 9, 0, 0, 0, domain.HospitalLocation)
	_, e := s.SaveDoctor(ctx, actors[0], domain.Doctor{ID: "doctor", Department: "General", SlotMinutes: 30, Hours: []domain.DoctorHours{{Weekday: int(day.Weekday()), StartMinute: 540, EndMinute: 1020}}})
	if e != nil {
		t.Fatal(e)
	}
	in := domain.AppointmentInput{PatientID: patients[0].ID, DoctorID: "doctor", StartsAt: start}
	app, e := s.Book(ctx, actors[0], in, "reschedule-booking-key")
	if e != nil {
		t.Fatal(e)
	}
	absence := domain.AbsenceInput{DoctorID: "doctor", StartsAt: start, EndsAt: start.Add(time.Hour), Reason: "Leave"}
	if _, e = s.CreateAbsence(ctx, actors[0], absence); !errors.Is(e, domain.ErrStale) {
		t.Fatal("absence over booking", e)
	}
	absence.StartsAt = start.Add(2 * time.Hour)
	absence.EndsAt = start.Add(3 * time.Hour)
	if _, e = s.CreateAbsence(ctx, actors[2], absence); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("other doctor absence", e)
	}
	off, e := s.CreateAbsence(ctx, actors[1], absence)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Reschedule(ctx, actors[0], app.ID, domain.RescheduleInput{StartsAt: absence.StartsAt, Version: 1, Reason: "Change"}); !errors.Is(e, domain.ErrStale) {
		t.Fatal("rescheduled into absence", e)
	}
	slots, e := s.Slots(ctx, actors[0], "doctor", day.Format("2006-01-02"))
	if e != nil {
		t.Fatal(e)
	}
	for _, slot := range slots {
		if !slot.Before(absence.StartsAt) && slot.Before(absence.EndsAt) {
			t.Fatal("absence slot offered")
		}
	}
	if _, e = s.CancelAbsence(ctx, actors[2], off.ID, domain.AbsenceCancel{Version: 1, Reason: "Forged"}); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("other doctor cancelled absence", e)
	}
	if _, e = s.CancelAbsence(ctx, actors[1], off.ID, domain.AbsenceCancel{Version: 1, Reason: "Leave cancelled"}); e != nil {
		t.Fatal(e)
	}
	if _, e = s.CancelAbsence(ctx, actors[1], off.ID, domain.AbsenceCancel{Version: 1, Reason: "Again"}); !errors.Is(e, domain.ErrStale) {
		t.Fatal("stale absence cancellation", e)
	}
	if _, e = s.Reschedule(ctx, actors[2], app.ID, domain.RescheduleInput{StartsAt: absence.StartsAt, Version: 1, Reason: "Change"}); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("other doctor rescheduled", e)
	}
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, errs[n] = s.Reschedule(ctx, actors[0], app.ID, domain.RescheduleInput{StartsAt: absence.StartsAt, Version: 1, Reason: "Patient requested"})
		}(n)
	}
	wg.Wait()
	if !((errs[0] == nil && errors.Is(errs[1], domain.ErrStale)) || (errs[1] == nil && errors.Is(errs[0], domain.ErrStale))) {
		t.Fatal("reschedule race", errs)
	}
	retry, e := s.Book(ctx, actors[0], in, "reschedule-booking-key")
	if e != nil || retry.ID != app.ID || !retry.StartsAt.Equal(absence.StartsAt) {
		t.Fatal("original booking retry", retry, e)
	}
	history, e := s.RescheduleHistory(ctx, actors[3], app.ID, 1)
	if e != nil || len(history) != 1 || !history[0].FromStart.Equal(start) {
		t.Fatal("scoped reschedule history", history, e)
	}
	if _, e = s.RescheduleHistory(ctx, actors[2], app.ID, 1); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("history scope", e)
	}
	if _, e = db.Exec(ctx, `DELETE FROM appointment_reschedule WHERE appointment_id=$1`, app.ID); e == nil {
		t.Fatal("reschedule history deleted")
	}
	if _, e = db.Exec(ctx, `UPDATE appointment SET original_starts_at=starts_at WHERE id=$1`, app.ID); e == nil {
		t.Fatal("original booking changed")
	}
	// A booking and leave request for the same free time serialize on the doctor row.
	raceStart := start.Add(4 * time.Hour)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, errs[0] = s.Book(ctx, actors[0], domain.AppointmentInput{PatientID: patients[1].ID, DoctorID: "doctor", StartsAt: raceStart}, "absence-booking-race")
	}()
	go func() {
		defer wg.Done()
		_, errs[1] = s.CreateAbsence(ctx, actors[0], domain.AbsenceInput{DoctorID: "doctor", StartsAt: raceStart, EndsAt: raceStart.Add(time.Hour), Reason: "Leave"})
	}()
	wg.Wait()
	if !((errs[0] == nil && errors.Is(errs[1], domain.ErrStale)) || (errs[1] == nil && errors.Is(errs[0], domain.ErrStale))) {
		t.Fatal("absence booking race", errs)
	}
	cancellation := domain.AppointmentChange{Status: "cancelled", Version: retry.Version, Reason: "Patient requested another day"}
	wg.Add(2)
	for n := 0; n < 2; n++ {
		go func(index int) { defer wg.Done(); _, errs[index] = s.Change(ctx, actors[3], app.ID, cancellation) }(n)
	}
	wg.Wait()
	if !((errs[0] == nil && errors.Is(errs[1], domain.ErrStale)) || (errs[1] == nil && errors.Is(errs[0], domain.ErrStale))) {
		t.Fatal("cancellation race", errs)
	}
	statusHistory, e := s.StatusHistory(ctx, actors[3], app.ID, 1)
	if e != nil || len(statusHistory) != 1 || statusHistory[0].Reason != cancellation.Reason || statusHistory[0].PreviousStatus != "booked" || statusHistory[0].NextStatus != "cancelled" || statusHistory[0].ActorID != "patient" {
		t.Fatal(statusHistory, e)
	}
	if _, e = s.StatusHistory(ctx, actors[2], app.ID, 1); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal(e)
	}
	if _, e = db.Exec(ctx, `DELETE FROM appointment_status_event WHERE appointment_id=$1`, app.ID); e == nil {
		t.Fatal("status history deleted")
	}
	if _, e = s.Change(ctx, actors[0], app.ID, domain.AppointmentChange{Status: "cancelled", Version: 1, Reason: strings.Repeat("x", 1001)}); !errors.Is(e, domain.ErrValidation) {
		t.Fatal(e)
	}
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		who := strings.TrimPrefix(r.Header.Get("Cookie"), "session=")
		fmt.Fprintf(w, `{"user":{"id":%q},"session":{"userId":%q,"expiresAt":%q}}`, who, who, time.Now().Add(time.Hour).Format(time.RFC3339))
	}))
	defer auth.Close()
	handler := httpapi.Server{Scheduling: s, Actors: store, AuthURL: auth.URL, Origin: "http://hospital.test", Client: auth.Client()}.Handler()
	for _, tc := range []struct {
		method, path, actor, body string
		want                      int
	}{
		{"GET", "/v1/appointments/" + app.ID + "/status-history", "patient", "", 200},
		{"GET", "/v1/appointments/" + app.ID + "/status-history", "other-doctor", "", 404},
		{"GET", "/v1/doctor-absences?doctorId=doctor", "doctor", "", 200},
		{"GET", "/v1/doctor-absences?doctorId=doctor", "patient", "", 403},
		{"GET", "/v1/appointments/" + app.ID + "/reschedule-history", "patient", "", 200},
		{"GET", "/v1/appointments/" + app.ID + "/reschedule-history", "other-doctor", "", 404},
		{"POST", "/v1/appointments/" + app.ID + "/reschedule", "admin", `{"unexpected":true}`, 400},
		{"POST", "/v1/doctor-absences", "admin", `{}`, 422},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Cookie", "session="+tc.actor)
		req.Header.Set("Origin", "http://hospital.test")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Fatalf("scheduling HTTP wanted %d got %d: %s", tc.want, w.Code, w.Body.String())
		}
	}

}
