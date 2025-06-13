package types

type TicketEntry struct {
	Type    string `json:"type"`
	Ordem   int32  `json:"ordem"`
	Subject string `json:"subject"`
	Poster  string `json:"poster"`
	Body    string `json:"body"`
}

type TicketsExportedData map[string][]TicketEntry
