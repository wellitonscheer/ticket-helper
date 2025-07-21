package pgvecodel

import "github.com/pgvector/pgvector-go"

type Ticket struct {
	Id        int
	Type      string
	TicketId  int
	Subject   string
	Ordem     int
	Poster    string
	Body      string
	Embedding pgvector.Vector
}

func (t *Ticket) IsEmpty() bool {
	return t.Type == "" && t.TicketId == 0 && t.Subject == "" && t.Ordem == 0 && len(t.Embedding.Slice()) == 0
}
