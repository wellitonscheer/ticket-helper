package pgvecervi

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/wellitonscheer/ticket-helper/internal/client"
	appContext "github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/pgvec/pgvecodel"
)

const (
	blackLimitSimilaritySearch int = 10
)

type BlackEntriesService struct {
	Conn   *pgxpool.Pool
	AppCtx appContext.AppContext
}

func NewBlackEntriesService(appCtx appContext.AppContext) BlackEntriesService {
	return BlackEntriesService{
		Conn:   appCtx.PGVec,
		AppCtx: appCtx,
	}
}

func (blk BlackEntriesService) Create(entry pgvecodel.BlackEntry) error {
	ctx := context.Background()
	_, err := blk.Conn.Exec(
		ctx,
		`
			INSERT INTO black_entries (content, embedding)
			VALUES ($1, $2);
		`,
		entry.Content, entry.Embedding,
	)
	if err != nil {
		return fmt.Errorf("failed to create new black entry (entry=%+v): %v", entry, err)
	}

	return nil
}

func (blk BlackEntriesService) CreateFromContent(content string) error {
	embedding, err := client.GetSingleTextEmbedding(blk.AppCtx, content)
	if err != nil {
		return fmt.Errorf("failed to get content embeddings (content=%s): %v", content, err)
	}

	entryToCreate := pgvecodel.BlackEntry{
		Content:   content,
		Embedding: pgvector.NewVector(embedding),
	}

	return blk.Create(entryToCreate)
}

func (blk BlackEntriesService) GetByEntryId(entryId int) pgvecodel.BlackEntry {
	var entry pgvecodel.BlackEntry

	sqlStm := "SELECT * FROM black_entries WHERE id = $1"
	blk.Conn.QueryRow(context.Background(), sqlStm, entryId).Scan(&entry.Id, &entry.Content, &entry.Embedding)

	return entry
}

func (blk BlackEntriesService) SearchSimilarByEmbed(embed []float32) ([]pgvecodel.BlackEntrySimilaritySearch, error) {
	var entries []pgvecodel.BlackEntrySimilaritySearch

	sqlStm := "SELECT id, content, 1 - (embedding <=> $1) AS distance FROM black_entries ORDER BY distance DESC LIMIT $2;"
	err := pgxscan.Select(context.Background(), blk.Conn, &entries, sqlStm, pgvector.NewVector(embed), blackLimitSimilaritySearch)
	if err != nil {
		return entries, fmt.Errorf("failed to searchh black ticket entries by embed (embed=%v): %v", embed, err)
	}

	return entries, nil
}

func (blk BlackEntriesService) SearchSimilarByText(text string) ([]pgvecodel.BlackEntrySimilaritySearch, error) {
	embed, err := client.GetSingleTextEmbedding(blk.AppCtx, text)
	if err != nil {
		return []pgvecodel.BlackEntrySimilaritySearch{}, fmt.Errorf("failed to get text embeddings for the similarity search (text='%s'): %v", text, err)
	}

	return blk.SearchSimilarByEmbed(embed)
}
