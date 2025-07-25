package pgvecervi

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/wellitonscheer/ticket-helper/internal/client"
	appContext "github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/pgvec/pgvecodel"
	"github.com/wellitonscheer/ticket-helper/internal/types"
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
	embedInput := types.ClientEmbeddingInputs{
		Inputs: []string{content},
	}
	embeddings, err := client.GetTextEmbeddings(blk.AppCtx, &embedInput)
	if err != nil {
		return fmt.Errorf("failed to get content embeddings (content=%s): %v", content, err)
	}

	firstEmbedding := (*embeddings)[0]
	if len(firstEmbedding) == 0 {
		return fmt.Errorf("no embedding returned to the content provided (embedInput=%+v, embeddings=%+v)", embedInput, embeddings)
	}

	entryToCreate := pgvecodel.BlackEntry{
		Content:   content,
		Embedding: pgvector.NewVector(firstEmbedding),
	}

	return blk.Create(entryToCreate)
}
