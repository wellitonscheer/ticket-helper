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
	migrationFolder  string = "./internal/database/pgvec/migrations"
	expTicketsPath   string = "./data_source/tickets.json"
	logFilePath      string = "./internal/database/pgvec/logs.txt"
	blackEntriesPath string = "./data_source/black_entries_content.json"
)

type InsertPGVectorData struct {
	AppCtx     appContext.AppContext
	Cleaner    utils.EntryCleaner
	Logger     func(text string)
	EntryServi pgvecervi.TicketEntriesService
	BlackServi pgvecervi.BlackEntriesService
}

func NewInsertPGVectorData(appCtx appContext.AppContext, log func(text string)) *InsertPGVectorData {
	entryCleaner := utils.NewEntryCleaner()

	entryServi := pgvecervi.NewTicketEntriesService(appCtx)

	blackServi := pgvecervi.NewBlackEntriesService(appCtx)

	return &InsertPGVectorData{
		AppCtx:     appCtx,
		Cleaner:    entryCleaner,
		Logger:     log,
		EntryServi: entryServi,
		BlackServi: blackServi,
	}
}

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

	InsertData(appCtx)
}

func InsertData(appCtx appContext.AppContext) {
	logFile := OpenLogFile(logFilePath)
	defer logFile.Close()

	log := func(text string) {
		Log(logFile, text)
	}

	insertData := NewInsertPGVectorData(appCtx, log)

	insertData.InsertBlackEntries()

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

	// for range ticketEntries.Data {
	for _, entry := range ticketEntries.Data {
		insertData.InsertTicketEntries(entry)
	}
}

func (d *InsertPGVectorData) InsertBlackEntries() {
	blackContent, err := os.ReadFile(blackEntriesPath)
	if err != nil {
		fmt.Printf("failed to read black entries file path\n")
		panic(err)
	}

	var blackEntries types.BlackEntriesContent
	err = json.Unmarshal(blackContent, &blackEntries)
	if err != nil {
		fmt.Printf("failed to unmarshal black entries\n")
		panic(err)
	}

	for _, entry := range blackEntries {
		storedEntry := d.BlackServi.GetByEntryId(entry.Id)
		if !storedEntry.IsEmpty() {
			d.Logger(fmt.Sprintf("INFOB: black entry already in database (entryID=%d)", entry.Id))
			continue
		}

		err = d.BlackServi.CreateFromContent(entry.Content)
		if err != nil {
			d.Logger(fmt.Sprintf("ERRORB: failed to create new black entry with CreateFromContent (entry=%+v): %v", entry, err))
			continue
		}

		d.Logger(fmt.Sprintf("INFOB: black entry inserted (entryID=%d)", entry.Id))
	}
}

func (d *InsertPGVectorData) InsertTicketEntries(entry types.TicketEntry) {
	if entry.Body == "" {
		d.Logger(fmt.Sprintf("INFOE: empty entry body (entryTicketID=%d, entryTicketOrdem=%d)", entry.TicketId, entry.Ordem))
		return
	}

	storedTicket, err := d.EntryServi.GetUniqueByTicketIdAndOrdem(entry.TicketId, entry.Ordem)
	if err != nil {
		d.Logger(fmt.Sprintf("ERRORE: failed to get ticket entry by id and ordem (ticketId=%d, ordem=%d): %v", entry.TicketId, entry.Ordem, err))
		return
	}

	if !storedTicket.IsEmpty() {
		d.Logger(fmt.Sprintf("INFOE: ticket already in database (storedEntryId=%d, storedTicketID=%d, storedTicketOrdem=%d)", storedTicket.Id, storedTicket.TicketId, storedTicket.Ordem))
		return
	}

	cleanBody := d.Cleaner.Clean(entry.Body)
	if cleanBody == "" {
		d.Logger(fmt.Sprintf("INFOE: empty cleaned body (entryTicketID=%d, entryTicketOrdem=%d)", entry.TicketId, entry.Ordem))
		return
	}

	embedding, err := client.GetSingleTextEmbedding(d.AppCtx, cleanBody)
	if err != nil {
		d.Logger(fmt.Sprintf("ERRORE: failed to get entry body embedding (cleanBody=%+v): %v", cleanBody, err))
		return
	}

	similarBlackEntries, err := d.BlackServi.SearchSimilarByEmbed(embedding)
	if err != nil {
		d.Logger(fmt.Sprintf("ERRORE: failed to get similar black entries to the ticket entry (entryTicketID=%d, entryTicketOrdem=%d): %v", entry.TicketId, entry.Ordem, err))
	} else if len(similarBlackEntries) > 0 {
		if similarBlackEntries[0].Distance > float32(0.8) {
			d.Logger(fmt.Sprintf("INFOE: entry is similar to a black entry (entryTicketID=%d, entryTicketOrdem=%d, blackEntryId=%d)", entry.TicketId, entry.Ordem, similarBlackEntries[0].Id))
			return
		}
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

	err = d.EntryServi.Create(ticket)
	if err != nil {
		d.Logger(fmt.Sprintf("ERRORE: failed to create new ticket (ticket=%+v): %v", ticket, err))
		return
	}

	d.Logger(fmt.Sprintf("INFOE: ticket inserted (ticketId=%d, ticketOrdem=%d)", ticket.TicketId, ticket.Ordem))
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
