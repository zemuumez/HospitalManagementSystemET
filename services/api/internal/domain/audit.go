package domain

import "time"

type AuditEvent struct {
	ID         int64     `json:"id"`
	ActorID    string    `json:"actorId"`
	Action     string    `json:"action"`
	ResourceID string    `json:"resourceId"`
	CreatedAt  time.Time `json:"createdAt"`
}
type AuditFilter struct {
	Before                      int64
	ActorID, Action, ResourceID string
}
