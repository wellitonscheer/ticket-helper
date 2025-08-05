package types

import "github.com/wellitonscheer/ticket-helper/internal/database/pgvec/pgvecodel"

type TicketChunkGetInputFilters struct {
	Columns []string
	Values  []any
}

func (fil TicketChunkGetInputFilters) IsValid() bool {
	if len(fil.Columns) == 0 {
		return false
	}

	if len(fil.Columns) != len(fil.Values) {
		return false
	}

	for _, col := range fil.Columns {
		if !pgvecodel.IsValidTicketChunkColumn(col) {
			return false
		}
	}

	return true
}

type SearchComputeScoreInput struct {
	Search        string
	Limit         int
	RelevantScore float32
}
