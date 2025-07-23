package types

type TicketEntry struct {
	Type     string `json:"type"`
	TicketId int    `json:"ticket_id"`
	Subject  string `json:"subject"`
	Ordem    int    `json:"ordem"`
	Poster   string `json:"poster"`
	Body     string `json:"body"`
}

type ExportedTickets struct {
	Data []TicketEntry `json:"data"`
}
