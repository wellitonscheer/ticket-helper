package pgvecodel

import "github.com/pgvector/pgvector-go"

type TicketChunk struct {
	Id        int
	Type      string
	TicketId  int
	Subject   string
	Ordem     int
	Poster    string
	Chunk     string
	Embedding pgvector.Vector
}

func (t *TicketChunk) IsEmpty() bool {
	// id is autoincrement and begins in 1, will never be 0
	return t.Id == 0
}

func IsValidTicketChunkColumn(col string) bool {
	allowedColumns := map[string]bool{
		"id":        true,
		"type":      true,
		"ticket_id": true,
		"subject":   true,
		"ordem":     true,
		"poster":    true,
		"chunk":     true,
		"embedding": true,
	}

	return allowedColumns[col]
}

type TicketChunkSimilaritySearch struct {
	Id       int
	Type     string
	TicketId int
	Subject  string
	Ordem    int
	Poster   string
	Chunk    string
	Distance float32
}

func (t *TicketChunkSimilaritySearch) IsEmpty() bool {
	// id is autoincrement and begins in 1, will never be 0
	return t.Id == 0
}
