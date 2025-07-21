package pgvec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/wellitonscheer/ticket-helper/internal/client"
	appContext "github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/pgvec/pgvecervi"
	"github.com/wellitonscheer/ticket-helper/internal/database/pgvec/pgvecodel"
	"github.com/wellitonscheer/ticket-helper/internal/types"
)

const (
	migrationFolder string = "./internal/database/pgvec/migrations"
	expTicketsPath  string = "./data_source/tickets.json"
)

func InitiatePGVec(appCtx appContext.AppContext) {
	fmt.Println("\nInitiating PGVector migrations...\n")

	migrations, err := os.ReadDir(migrationFolder)
	if err != nil {
		fmt.Printf("error to read migration folder (entries read=%d)\n", len(migrations))
		panic(err)
	}

	for i, entry := range migrations {
		fmt.Printf("\nfile number=%d file name = %s \n", i, entry.Name())

		fileMigration, err := os.ReadFile(fmt.Sprintf("%s/%s", migrationFolder, entry.Name()))
		if err != nil {
			fmt.Println("error to read migration file")
			panic(err)
		}

		fmt.Printf("\nexecuting migration: \n\n %s \n", string(fileMigration))

		_, err = appCtx.PGVec.Exec(context.Background(), string(fileMigration))
		if err != nil {
			fmt.Println("error to execute migration")
			panic(err)
		}
	}

	fmt.Println("\nPGVector migrations applied\n")

	InsertTickets(appCtx)
}

func InsertTickets(appCtx appContext.AppContext) {
	tickets, err := os.ReadFile(expTicketsPath)
	if err != nil {
		fmt.Printf("error to read exported tickets file (path=%s)\n", expTicketsPath)
		panic(err)
	}

	var ticketEntries types.ExportedTickets
	err = json.Unmarshal(tickets, &ticketEntries)
	if err != nil {
		fmt.Printf("error to unmarshal exported tickets\n")
		panic(err)
	}

	ticketServi := pgvecervi.NewPGTicketServices(appCtx)

	for _, entry := range ticketEntries.Data {
		embedInputs := types.Inputs{
			Inputs: []string{entry.Body},
		}

		embeddings, err := client.GetTextEmbeddings(appCtx, &embedInputs)
		if err != nil {
			fmt.Printf("failed to get entry body embeddings (entry=%+v)\n", embedInputs)
			panic(err)
		}

		if len(*embeddings) == 0 {
			fmt.Printf("embedding has returned no value (embeddings=\n%+v\n)\n", embeddings)
			panic("")
		}

		ticket := pgvecodel.Ticket{
			Type:       entry.Type,
			TicketId:   entry.TicketId,
			Subject:    entry.Subject,
			Ordem:      entry.Ordem,
			Poster:     entry.Poster,
			Body:       entry.Body,
			Embeddings: (*embeddings)[0],
		}

		err = ticketServi.Create(ticket)
		if err != nil {
			fmt.Printf("failed to create new ticket\n")
			panic(err)
		}
	}
}
