package pgvecervi

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
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
		ticket.Type, ticket.TicketId, ticket.Subject, ticket.Ordem, ticket.Poster, ticket.Body, pgvector.NewVector(ticket.Embeddings),
	)
	if err != nil {
		return fmt.Errorf("failed to create new ticket (ticket=%+v): %v", ticket, err)
	}

	return nil
}
