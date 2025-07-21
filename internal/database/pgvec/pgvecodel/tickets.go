package pgvecodel

type Ticket struct {
	Type       string
	TicketId   int
	Subject    string
	Ordem      int
	Poster     string
	Body       string
	Embeddings []float32
}
