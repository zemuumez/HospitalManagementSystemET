package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) clinical(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	page := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		page, _ = strconv.Atoi(raw)
	}
	switch {
	case r.URL.Path == "/v1/bed-types" && r.Method == "GET":
		out, e := s.Clinical.BedTypes(r.Context(), a, page)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, map[string]any{"bedTypes": out, "page": page, "pageSize": 25})
		}
	case (r.URL.Path == "/v1/bed-types" && r.Method == "POST") || (strings.HasPrefix(r.URL.Path, "/v1/bed-types/") && r.Method == "PATCH"):
		var i domain.BedTypeInput
		if !decode(w, r, &i) {
			return true
		}
		id := ""
		if r.Method == "PATCH" {
			id = strings.TrimPrefix(r.URL.Path, "/v1/bed-types/")
		}
		out, e := s.Clinical.SaveBedType(r.Context(), a, id, i)
		if e != nil {
			fail(w, e)
		} else {
			status := 200
			if id == "" {
				status = 201
			}
			write(w, status, out)
		}

	case strings.HasPrefix(r.URL.Path, "/v1/beds/") && r.Method == "PATCH":
		id := strings.TrimPrefix(r.URL.Path, "/v1/beds/")
		var i domain.BedStateInput
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.Clinical.SetBedState(r.Context(), a, id, i)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, out)
		}
	case r.URL.Path == "/v1/beds" && r.Method == "GET":
		out, e := s.Clinical.Beds(r.Context(), a, page)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, map[string]any{"beds": out, "page": page, "pageSize": 25})
		}
	case r.URL.Path == "/v1/beds" && r.Method == "POST":
		var i domain.BedInput
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.Clinical.CreateBed(r.Context(), a, i)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 201, out)
		}
	case r.URL.Path == "/v1/cases" && r.Method == "GET":
		out, e := s.Clinical.Cases(r.Context(), a, page)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, map[string]any{"cases": out, "page": page, "pageSize": 25})
		}
	case r.URL.Path == "/v1/cases" && r.Method == "POST":
		var i domain.CaseInput
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.Clinical.CreateCase(r.Context(), a, i)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 201, out)
		}
	case r.URL.Path == "/v1/encounters" && r.Method == "GET":
		out, e := s.Clinical.Encounters(r.Context(), a, r.URL.Query().Get("kind"), page)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, map[string]any{"encounters": out, "page": page, "pageSize": 25})
		}
	case r.URL.Path == "/v1/encounters" && r.Method == "POST":
		var i domain.EncounterInput
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.Clinical.Admit(r.Context(), a, i, r.Header.Get("Idempotency-Key"))
		if e != nil {
			fail(w, e)
		} else {
			write(w, 201, out)
		}
	case strings.HasPrefix(r.URL.Path, "/v1/encounters/"):
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/encounters/"), "/")
		if len(parts) != 2 {
			return false
		}
		id := parts[0]
		switch {
		case parts[1] == "addenda" && r.Method == "GET":
			out, e := s.Clinical.Addenda(r.Context(), a, id, page)
			if e != nil {
				fail(w, e)
			} else {
				write(w, 200, map[string]any{"addenda": out, "page": page, "pageSize": 25})
			}
		case parts[1] == "addenda" && r.Method == "POST":
			var i domain.AddendumInput
			if !decode(w, r, &i) {
				return true
			}
			out, e := s.Clinical.AddAddendum(r.Context(), a, id, i, r.Header.Get("Idempotency-Key"))
			if e != nil {
				fail(w, e)
			} else {
				write(w, 201, out)
			}
		case parts[1] == "bed-history" && r.Method == "GET":
			out, e := s.Clinical.BedHistory(r.Context(), a, id, page)
			if e != nil {
				fail(w, e)
			} else {
				write(w, 200, map[string]any{"events": out, "page": page, "pageSize": 25})
			}
		case parts[1] == "transfer" && r.Method == "POST":
			var i domain.BedTransfer
			if !decode(w, r, &i) {
				return true
			}
			out, e := s.Clinical.TransferBed(r.Context(), a, id, i)
			if e != nil {
				fail(w, e)
			} else {
				write(w, 200, out)
			}
		case parts[1] == "discharge" && r.Method == "POST":
			var i domain.Discharge
			if !decode(w, r, &i) {
				return true
			}
			out, e := s.Clinical.Discharge(r.Context(), a, id, i)
			if e != nil {
				fail(w, e)
			} else {
				write(w, 200, out)
			}
		case parts[1] == "notes" && r.Method == "GET":
			out, e := s.Clinical.Notes(r.Context(), a, id)
			if e != nil {
				fail(w, e)
			} else {
				write(w, 200, map[string]any{"notes": out})
			}
		case parts[1] == "notes" && r.Method == "POST":
			var i domain.NoteInput
			if !decode(w, r, &i) {
				return true
			}
			out, e := s.Clinical.SignNote(r.Context(), a, id, i, r.Header.Get("Idempotency-Key"))
			if e != nil {
				fail(w, e)
			} else {
				write(w, 201, out)
			}
		default:
			return false
		}
	default:
		return false
	}
	return true
}
