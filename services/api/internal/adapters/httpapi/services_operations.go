package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) servicesOperations(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	// Charge Categories
	if strings.HasPrefix(r.URL.Path, "/v1/charge-categories") {
		switch {
		case r.URL.Path == "/v1/charge-categories" && r.Method == "GET":
			chargeType, _ := strconv.Atoi(r.URL.Query().Get("type"))
			list, err := s.ServicesOperations.ChargeCategories(r.Context(), a, chargeType)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"charge_categories": list})
			return true

		case r.URL.Path == "/v1/charge-categories" && r.Method == "POST":
			var in domain.ChargeCategoryInput
			if !decode(w, r, &in) {
				return true
			}
			cat, err := s.ServicesOperations.CreateChargeCategory(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, cat)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/charge-categories/") && r.Method == "GET":
			id := strings.TrimPrefix(r.URL.Path, "/v1/charge-categories/")
			cat, err := s.ServicesOperations.ChargeCategory(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, cat)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/charge-categories/") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/charge-categories/")
			var in domain.ChargeCategoryInput
			if !decode(w, r, &in) {
				return true
			}
			cat, err := s.ServicesOperations.UpdateChargeCategory(r.Context(), a, id, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, cat)
			return true
		}
	}

	// Charges
	if strings.HasPrefix(r.URL.Path, "/v1/charges") {
		switch {
		case r.URL.Path == "/v1/charges" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			categoryID := r.URL.Query().Get("category_id")
			list, total, err := s.ServicesOperations.Charges(r.Context(), a, page, categoryID)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"charges": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/charges" && r.Method == "POST":
			var in domain.HospitalChargeInput
			if !decode(w, r, &in) {
				return true
			}
			ch, err := s.ServicesOperations.CreateCharge(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, ch)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/charges/") && r.Method == "GET":
			id := strings.TrimPrefix(r.URL.Path, "/v1/charges/")
			ch, err := s.ServicesOperations.Charge(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, ch)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/charges/") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/charges/")
			var in domain.HospitalChargeInput
			if !decode(w, r, &in) {
				return true
			}
			ch, err := s.ServicesOperations.UpdateCharge(r.Context(), a, id, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, ch)
			return true
		}
	}

	// Services
	if strings.HasPrefix(r.URL.Path, "/v1/services") {
		switch {
		case r.URL.Path == "/v1/services" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			if page < 1 {
				page = 1
			}
			limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
			if limit < 1 || limit > 100 {
				limit = 25
			}
			search := r.URL.Query().Get("search")
			var status *int
			if qStatus := r.URL.Query().Get("status"); qStatus != "" {
				st, _ := strconv.Atoi(qStatus)
				status = &st
			}
			list, total, err := s.ServicesOperations.Services(r.Context(), a, page, limit, status, search)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"services": list, "total": total, "page": page, "limit": limit})
			return true

		case r.URL.Path == "/v1/services" && r.Method == "POST":
			var in domain.HospitalServiceInput
			if !decode(w, r, &in) {
				return true
			}
			srv, err := s.ServicesOperations.CreateService(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, srv)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/services/") && r.Method == "GET":
			id := strings.TrimPrefix(r.URL.Path, "/v1/services/")
			srv, err := s.ServicesOperations.Service(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, srv)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/services/") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/services/")
			var in domain.HospitalServiceInput
			if !decode(w, r, &in) {
				return true
			}
			srv, err := s.ServicesOperations.UpdateService(r.Context(), a, id, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, srv)
			return true
		}
	}

	// Operation Categories
	if strings.HasPrefix(r.URL.Path, "/v1/operation-categories") {
		switch {
		case r.URL.Path == "/v1/operation-categories" && r.Method == "GET":
			list, err := s.ServicesOperations.OperationCategories(r.Context(), a)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"operation_categories": list})
			return true
		case r.URL.Path == "/v1/operation-categories" && r.Method == "POST":
			var in domain.OperationCategoryInput
			if !decode(w, r, &in) {
				return true
			}
			cat, err := s.ServicesOperations.CreateOperationCategory(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, cat)
			return true
		case strings.HasPrefix(r.URL.Path, "/v1/operation-categories/") && r.Method == "DELETE":
			id := strings.TrimPrefix(r.URL.Path, "/v1/operation-categories/")
			if err := s.ServicesOperations.DeleteOperationCategory(r.Context(), a, id); err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]bool{"deleted": true})
			return true
		}
	}

	// Operations
	if strings.HasPrefix(r.URL.Path, "/v1/operations") {
		switch {
		case r.URL.Path == "/v1/operations" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			categoryID := r.URL.Query().Get("category_id")
			list, total, err := s.ServicesOperations.Operations(r.Context(), a, page, categoryID)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"operations": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/operations" && r.Method == "POST":
			var in domain.HospitalOperationInput
			if !decode(w, r, &in) {
				return true
			}
			op, err := s.ServicesOperations.CreateOperation(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, op)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/operations/") && r.Method == "GET":
			id := strings.TrimPrefix(r.URL.Path, "/v1/operations/")
			op, err := s.ServicesOperations.Operation(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, op)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/operations/") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/operations/")
			var in domain.HospitalOperationInput
			if !decode(w, r, &in) {
				return true
			}
			op, err := s.ServicesOperations.UpdateOperation(r.Context(), a, id, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, op)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/operations/") && r.Method == "DELETE":
			id := strings.TrimPrefix(r.URL.Path, "/v1/operations/")
			if err := s.ServicesOperations.DeleteOperation(r.Context(), a, id); err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]bool{"deleted": true})
			return true
		}
	}

	// Custom Fields
	if strings.HasPrefix(r.URL.Path, "/v1/custom-fields") {
		switch {
		case r.URL.Path == "/v1/custom-fields" && r.Method == "GET":
			moduleName := r.URL.Query().Get("module")
			list, err := s.ServicesOperations.CustomFields(r.Context(), a, moduleName)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"custom_fields": list})
			return true

		case r.URL.Path == "/v1/custom-fields" && r.Method == "POST":
			var in domain.CustomFieldInput
			if !decode(w, r, &in) {
				return true
			}
			cf, err := s.ServicesOperations.CreateCustomField(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, cf)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/custom-fields/") && r.Method == "DELETE":
			id := strings.TrimPrefix(r.URL.Path, "/v1/custom-fields/")
			if err := s.ServicesOperations.DeleteCustomField(r.Context(), a, id); err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]bool{"deleted": true})
			return true
		}
	}

	// Module Settings
	if strings.HasPrefix(r.URL.Path, "/v1/modules-setting") || strings.HasPrefix(r.URL.Path, "/v1/module-settings") {
		switch {
		case (r.URL.Path == "/v1/modules-setting" || r.URL.Path == "/v1/module-settings") && r.Method == "GET":
			list, err := s.ServicesOperations.ModuleSettings(r.Context(), a)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"modules_setting": list, "module_settings": list})
			return true

		case (strings.HasPrefix(r.URL.Path, "/v1/modules-setting/") || strings.HasPrefix(r.URL.Path, "/v1/module-settings/")) && (r.Method == "PUT" || r.Method == "PATCH"):
			key := strings.TrimPrefix(r.URL.Path, "/v1/modules-setting/")
			key = strings.TrimPrefix(key, "/v1/module-settings/")
			var in domain.ModuleSettingUpdateInput
			if !decode(w, r, &in) {
				return true
			}
			ms, err := s.ServicesOperations.UpdateModuleSetting(r.Context(), a, key, in.IsActive)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, ms)
			return true
		}
	}

	return false
}
