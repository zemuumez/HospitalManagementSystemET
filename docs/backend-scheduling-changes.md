# Absences and appointment rescheduling

Source reviewed for this increment: Laravel `DoctorHoliday` and `HolidayRepository`. Original holidays are doctor/date/name records. The new API accepts timestamp ranges (including a whole EAT calendar day) to support holidays and partial-day leave. This is an intentional extension; original filters and frontend are deferred.

- `GET /v1/doctor-absences?doctorId=...&page=1`: admin or the named doctor; 25 records per page, including retained cancellations.
- `POST /v1/doctor-absences`: `{doctorId, startsAt, endsAt, reason}`. Admin or that doctor. Future range within the coming year; required reason up to 200 characters. Overlapping active absence or booked/arrived appointments gives STATE_CONFLICT. Doctor must be active and have a profile.
- `PATCH /v1/doctor-absences/{id}`: `{version, reason}` cancels the absence without deleting its original range/reason. The cancellation reason/actor/time are retained. Other doctors receive NOT_FOUND.
- `POST /v1/appointments/{id}/reschedule`: `{startsAt, version, reason}`. Admin, receptionist, assigned doctor or owning patient; future booked appointments only. The doctor remains the same. Server derives duration, validates working periods and absence/doctor/patient conflicts, records old/new times and increments version. Stale retries cannot create a second change. Original booking-key retries still resolve after a move.
- `GET /v1/appointments/{id}/reschedule-history?page=1`: same record scope as appointment reads; 25 immutable events per page.

Booking, leave and rescheduling take the same doctor-row lock. Rescheduling then locks the patient and appointment in that order. Existing status-change behavior remains. Cancelled absence no longer blocks slot discovery or booking. A move does not yet enqueue an SMS/email update; staff must communicate changes until the notification workflow is connected. Cross-doctor reassignment, reminder job cancellation/replacement and per-date schedule overrides remain open. No new UI integration was added.

Tests cover busy leave rejection, doctor ownership, cancellation versions, hidden unavailable slots, immutable original booking/history, concurrent rescheduling, competing absence/booking, patient history scope, strict HTTP inputs and role denial.
