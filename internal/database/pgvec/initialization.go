package pgvec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pgvector/pgvector-go"
	"github.com/wellitonscheer/ticket-helper/internal/client"
	appContext "github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/pgvec/pgvecervi"
	"github.com/wellitonscheer/ticket-helper/internal/database/pgvec/pgvecodel"
	"github.com/wellitonscheer/ticket-helper/internal/types"
	"github.com/wellitonscheer/ticket-helper/internal/utils"
)

const (
	migrationFolder string = "./internal/database/pgvec/migrations"
	expTicketsPath  string = "./data_source/tickets.json"
	logFilePath     string = "./internal/database/pgvec/logs.txt"
)

func RunMigrations(appCtx appContext.AppContext) {
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

	// InsertTickets(appCtx)
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

	logFile := OpenLogFile(logFilePath)
	defer logFile.Close()

	entryCleaner := utils.NewEntryCleaner()

	ticketServi := pgvecervi.NewTicketEntriesService(appCtx)

	for _, entry := range ticketEntries.Data {
		if entry.Body == "" {
			Log(logFile, fmt.Sprintf("INFO: empty entry body (entryTicketID=%d, entryTicketOrdem=%d)", entry.TicketId, entry.Ordem))
			continue
		}

		storedTicket, err := ticketServi.GetUniqueByTicketIdAndOrdem(entry.TicketId, entry.Ordem)
		if err != nil {
			Log(logFile, fmt.Sprintf("ERROR: failed to get ticket entry by id and ordem (ticketId=%d, ordem=%d): %v", entry.TicketId, entry.Ordem, err))
			continue
		}

		if !storedTicket.IsEmpty() {
			// already in database
			Log(logFile, fmt.Sprintf("INFO: ticket already in database (storedEntryId=%d, storedTicketID=%d, storedTicketOrdem=%d)", storedTicket.Id, storedTicket.TicketId, storedTicket.Ordem))
			continue
		}

		cleanBody := entryCleaner.Clean(entry.Body)
		if cleanBody == "" {
			Log(logFile, fmt.Sprintf("INFO: empty cleaned body (entryTicketID=%d, entryTicketOrdem=%d)", entry.TicketId, entry.Ordem))
			continue
		}

		embedding, err := client.GetSingleTextEmbedding(appCtx, cleanBody)
		if err != nil {
			Log(logFile, fmt.Sprintf("ERROR: failed to get entry body embedding (cleanBody=%+v): %v", cleanBody, err))
			continue
		}

		ticket := pgvecodel.TicketEntry{
			Type:      entry.Type,
			TicketId:  entry.TicketId,
			Subject:   entry.Subject,
			Ordem:     entry.Ordem,
			Poster:    entry.Poster,
			Body:      cleanBody,
			Embedding: pgvector.NewVector(embedding),
		}

		err = ticketServi.Create(ticket)
		if err != nil {
			Log(logFile, fmt.Sprintf("ERROR: failed to create new ticket (ticket=%+v): %v", ticket, err))
			continue
		}

		Log(logFile, fmt.Sprintf("INFO: ticket inserted (ticketId=%d, ticketOrdem=%d)", ticket.TicketId, ticket.Ordem))
	}
}

func OpenLogFile(path string) *os.File {
	logFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("failed to open log file (path=%s)", path)
		panic(err)
	}

	return logFile
}

func Log(file *os.File, log string) {
	info := fmt.Sprintf("%v: %s\n", time.Now(), log)

	_, err := file.Write([]byte(info))
	if err != nil {
		fmt.Printf("error to write into log file (info=%s): %v", info, err)
	}
}
