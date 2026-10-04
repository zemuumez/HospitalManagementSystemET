package domain

// MessageLease is internal worker data; never serialize recipient/body into logs.
type MessageLease struct {
	ID, Token, Channel, Recipient, Subject, Body string
}
