package pgvecervi

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	appContext "github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/pgvec/pgvecodel"
)

type TicketEntriesService struct {
	Conn   *pgxpool.Pool
	AppCtx appContext.AppContext
}

func NewPGTicketServices(appCtx appContext.AppContext) TicketEntriesService {
	return TicketEntriesService{
		Conn:   appCtx.PGVec,
		AppCtx: appCtx,
	}
}

func (tik TicketEntriesService) Create(ticket pgvecodel.TicketEntry) error {
	ctx := context.Background()
	_, err := tik.Conn.Exec(
		ctx,
		`
			INSERT INTO ticket_entries (type, ticket_id, subject, ordem, poster, body, embedding)
			VALUES ($1, $2, $3, $4, $5, $6, $7);
		`,
		ticket.Type, ticket.TicketId, ticket.Subject, ticket.Ordem, ticket.Poster, ticket.Body, ticket.Embedding,
	)
	if err != nil {
		return fmt.Errorf("failed to create new ticket (ticket=%+v): %v", ticket, err)
	}

	return nil
}

func (tik TicketEntriesService) GetByTicketId(ticketId int) ([]pgvecodel.TicketEntry, error) {
	var ticket []pgvecodel.TicketEntry

	err := pgxscan.Select(context.Background(), tik.Conn, &ticket, "SELECT * FROM ticket_entries WHERE ticket_id = $1", ticketId)
	if err != nil {
		return ticket, fmt.Errorf("failed to get ticket entries by id (ticketId=%d): %v", ticketId, err)
	}

	return ticket, nil
}

func (tik TicketEntriesService) GetUniqueByTicketIdAndOrdem(ticketId int, ordem int) (pgvecodel.TicketEntry, error) {
	var ticket []pgvecodel.TicketEntry

	err := pgxscan.Select(context.Background(), tik.Conn, &ticket, "SELECT * FROM ticket_entries WHERE ticket_id = $1 AND ordem = $2", ticketId, ordem)
	if err != nil {
		return pgvecodel.TicketEntry{}, fmt.Errorf("failed to get unique ticket entry by id and ordem (ticketId=%d): %v", ticketId, err)
	}

	if len(ticket) == 0 {
		return pgvecodel.TicketEntry{}, nil
	}

	return ticket[0], nil
}
