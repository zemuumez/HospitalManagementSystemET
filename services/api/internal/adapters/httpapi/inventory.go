package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) inventory(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	if !strings.HasPrefix(r.URL.Path, "/v1/inventory/") {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/inventory/"), "/")
	page := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		page, _ = strconv.Atoi(raw)
	}
	var out any
	var e error
	status := 200
	switch {
	case len(parts) == 1 && parts[0] == "categories" && r.Method == "GET":
		var data []domain.InventoryCategory
		data, e = s.Inventory.Categories(r.Context(), a, page)
		out = map[string]any{"categories": data}
	case parts[0] == "categories" && ((len(parts) == 1 && r.Method == "POST") || (len(parts) == 2 && r.Method == "PATCH")):
		var i domain.InventoryCategoryInput
		if !decode(w, r, &i) {
			return true
		}
		id := ""
		if len(parts) == 2 {
			id = parts[1]
		} else {
			status = 201
		}
		out, e = s.Inventory.SaveCategory(r.Context(), a, id, i)
	case len(parts) == 2 && parts[0] == "categories" && r.Method == "DELETE":
		id := parts[1]
		e = s.Inventory.DeleteCategory(r.Context(), a, id)
		out = map[string]bool{"deleted": true}
	case len(parts) == 1 && parts[0] == "items" && r.Method == "GET":
		var data []domain.InventoryItem
		data, e = s.Inventory.Items(r.Context(), a, r.URL.Query().Get("search"), r.URL.Query().Get("lowStock") == "true", page)
		out = map[string]any{"items": data}
	case parts[0] == "items" && ((len(parts) == 1 && r.Method == "POST") || (len(parts) == 2 && r.Method == "PATCH")):
		var i domain.InventoryItemInput
		if !decode(w, r, &i) {
			return true
		}
		id := ""
		if len(parts) == 2 {
			id = parts[1]
		} else {
			status = 201
		}
		out, e = s.Inventory.SaveItem(r.Context(), a, id, i)
	case len(parts) == 2 && parts[0] == "items" && r.Method == "DELETE":
		id := parts[1]
		e = s.Inventory.DeleteItem(r.Context(), a, id)
		out = map[string]bool{"deleted": true}
	case len(parts) == 1 && parts[0] == "movements" && r.Method == "POST":
		var i domain.InventoryMovementInput
		if !decode(w, r, &i) {
			return true
		}
		out, e = s.Inventory.Move(r.Context(), a, i, r.Header.Get("Idempotency-Key"))
		status = 201
	case len(parts) == 1 && parts[0] == "movements" && r.Method == "GET":
		var data []domain.InventoryMovement
		data, e = s.Inventory.Movements(r.Context(), a, r.URL.Query().Get("itemId"), page)
		out = map[string]any{"movements": data}
	default:
		return false
	}
	if e != nil {
		fail(w, e)
	} else {
		write(w, status, out)
	}
	return true
}
