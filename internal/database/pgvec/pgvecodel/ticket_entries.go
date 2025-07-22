package pgvecodel

import "github.com/pgvector/pgvector-go"

type TicketEntry struct {
	Id        int
	Type      string
	TicketId  int
	Subject   string
	Ordem     int
	Poster    string
	Body      string
	Embedding pgvector.Vector
}

func (t *TicketEntry) IsEmpty() bool {
	// id is autoincrement and begins in 1, will never be 0
	return t.Id == 0
}
