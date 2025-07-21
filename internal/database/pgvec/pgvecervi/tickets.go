package pgvecervi

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	appContext "github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/pgvec/pgvecodel"
)

type TicketServices struct {
	Conn   *pgxpool.Pool
	AppCtx appContext.AppContext
}

func NewPGTicketServices(appCtx appContext.AppContext) TicketServices {
	return TicketServices{
		Conn:   appCtx.PGVec,
		AppCtx: appCtx,
	}
}

func (tik TicketServices) Create(ticket pgvecodel.Ticket) error {
	ctx := context.Background()
	_, err := tik.Conn.Exec(
		ctx,
		`
			INSERT INTO tickets (type, ticket_id, subject, ordem, poster, body, embedding)
			VALUES ($1, $2, $3, $4, $5, $6, $7);
		`,
		ticket.Type, ticket.TicketId, ticket.Subject, ticket.Ordem, ticket.Poster, ticket.Body, ticket.Embedding,
	)
	if err != nil {
		return fmt.Errorf("failed to create new ticket (ticket=%+v): %v", ticket, err)
	}

	return nil
}

func (tik TicketServices) GetByTicketId(ticketId int) (pgvecodel.Ticket, error) {
	var ticket []pgvecodel.Ticket

	err := pgxscan.Select(context.Background(), tik.Conn, &ticket, "SELECT * FROM tickets WHERE ticket_id = $1", ticketId)
	if err != nil {
		return pgvecodel.Ticket{}, fmt.Errorf("failed to get ticket by id (ticketId=%d): %v", ticketId, err)
	}

	if len(ticket) == 0 {
		return pgvecodel.Ticket{}, nil
	}

	return ticket[0], nil
}
