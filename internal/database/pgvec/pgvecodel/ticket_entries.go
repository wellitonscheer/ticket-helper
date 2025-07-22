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
	return t.Id == 0
}
