package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) pharmacyBloodBank(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	// 1. Medicine Categories
	if strings.HasPrefix(r.URL.Path, "/v1/medicine-categories") {
		switch {
		case r.URL.Path == "/v1/medicine-categories" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			search := r.URL.Query().Get("search")
			list, total, err := s.PharmacyBloodBank.MedicineCategories(r.Context(), a, page, search)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"categories": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/medicine-categories" && r.Method == "POST":
			var in domain.MedicineCategoryInput
			if !decode(w, r, &in) {
				return true
			}
			cat, err := s.PharmacyBloodBank.CreateMedicineCategory(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, cat)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/medicine-categories/") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/medicine-categories/")
			var in domain.MedicineCategoryInput
			if !decode(w, r, &in) {
				return true
			}
			cat, err := s.PharmacyBloodBank.UpdateMedicineCategory(r.Context(), a, id, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, cat)
			return true
		}
	}

	// 2. Medicine Brands
	if strings.HasPrefix(r.URL.Path, "/v1/medicine-brands") {
		switch {
		case r.URL.Path == "/v1/medicine-brands" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			search := r.URL.Query().Get("search")
			list, total, err := s.PharmacyBloodBank.MedicineBrands(r.Context(), a, page, search)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"brands": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/medicine-brands" && r.Method == "POST":
			var in domain.MedicineBrandInput
			if !decode(w, r, &in) {
				return true
			}
			b, err := s.PharmacyBloodBank.CreateMedicineBrand(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, b)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/medicine-brands/") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/medicine-brands/")
			var in domain.MedicineBrandInput
			if !decode(w, r, &in) {
				return true
			}
			b, err := s.PharmacyBloodBank.UpdateMedicineBrand(r.Context(), a, id, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, b)
			return true
		}
	}

	// 3. Blood Bank
	if r.URL.Path == "/v1/blood-bank" && r.Method == "GET" {
		list, err := s.PharmacyBloodBank.BloodBank(r.Context(), a)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"blood_bank": list})
		return true
	}

	// 4. Blood Donors
	if strings.HasPrefix(r.URL.Path, "/v1/blood-donors") {
		switch {
		case r.URL.Path == "/v1/blood-donors" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			search := r.URL.Query().Get("search")
			list, total, err := s.PharmacyBloodBank.BloodDonors(r.Context(), a, page, search)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"donors": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/blood-donors" && r.Method == "POST":
			var in domain.BloodDonorInput
			if !decode(w, r, &in) {
				return true
			}
			d, err := s.PharmacyBloodBank.CreateBloodDonor(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, d)
			return true
		}
	}

	// 5. Blood Donations
	if strings.HasPrefix(r.URL.Path, "/v1/blood-donations") {
		switch {
		case r.URL.Path == "/v1/blood-donations" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			list, total, err := s.PharmacyBloodBank.BloodDonations(r.Context(), a, page)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"donations": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/blood-donations" && r.Method == "POST":
			var in domain.BloodDonationInput
			if !decode(w, r, &in) {
				return true
			}
			d, err := s.PharmacyBloodBank.RecordBloodDonation(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, d)
			return true
		}
	}

	// 6. Blood Issues
	if strings.HasPrefix(r.URL.Path, "/v1/blood-issues") {
		switch {
		case r.URL.Path == "/v1/blood-issues" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			list, total, err := s.PharmacyBloodBank.BloodIssues(r.Context(), a, page)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"issues": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/blood-issues" && r.Method == "POST":
			var in domain.BloodIssueInput
			if !decode(w, r, &in) {
				return true
			}
			bi, err := s.PharmacyBloodBank.CreateBloodIssue(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, bi)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/blood-issues/") && r.Method == "GET":
			id := strings.TrimPrefix(r.URL.Path, "/v1/blood-issues/")
			bi, err := s.PharmacyBloodBank.BloodIssue(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, bi)
			return true
		}
	}

	// 7. Prescriptions
	if strings.HasPrefix(r.URL.Path, "/v1/prescriptions") {
		switch {
		case r.URL.Path == "/v1/prescriptions" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			list, total, err := s.PharmacyBloodBank.Prescriptions(r.Context(), a, page)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"prescriptions": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/prescriptions" && r.Method == "POST":
			var in domain.PrescriptionInput
			if !decode(w, r, &in) {
				return true
			}
			p, err := s.PharmacyBloodBank.CreatePrescription(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, p)
			return true

		case strings.HasSuffix(r.URL.Path, "/status") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/prescriptions/")
			id = strings.TrimSuffix(id, "/status")
			var in struct {
				Status int `json:"status"`
			}
			if !decode(w, r, &in) {
				return true
			}
			p, err := s.PharmacyBloodBank.UpdatePrescriptionStatus(r.Context(), a, id, in.Status)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, p)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/prescriptions/") && r.Method == "GET":
			id := strings.TrimPrefix(r.URL.Path, "/v1/prescriptions/")
			p, err := s.PharmacyBloodBank.Prescription(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, p)
			return true
		}
	}

	return false
}
