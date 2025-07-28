package pgvecodel

import "github.com/pgvector/pgvector-go"

type BlackEntry struct {
	Id        int
	Content   string
	Embedding pgvector.Vector
}

func (b *BlackEntry) IsEmpty() bool {
	// id is autoincrement and begins in 1, will never be 0
	return b.Id == 0
}
