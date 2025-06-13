package types

type TicketSearchResult struct {
	TicketId string
	Score    float32
}

type TicketSearchResults = []TicketSearchResult
