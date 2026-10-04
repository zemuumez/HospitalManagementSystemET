package domain

import "time"

type AddendumInput struct {
	Kind   string `json:"kind"`
	NoteID string `json:"noteId"`
	Body   string `json:"body"`
	Reason string `json:"reason"`
}
type Addendum struct {
	AddendumInput
	ID          string    `json:"id"`
	EncounterID string    `json:"encounterId"`
	AuthorID    string    `json:"authorId"`
	SignedAt    time.Time `json:"signedAt"`
}

func (i *AddendumInput) Validate() error {
	if !validText(&i.Body, 1, 10000) || !validText(&i.Reason, 1, 1000) {
		return ErrValidation
	}
	if (i.Kind == "note" && UUIDPattern.MatchString(i.NoteID)) || (i.Kind == "discharge" && i.NoteID == "") {
		return nil
	}
	return ErrValidation
}
