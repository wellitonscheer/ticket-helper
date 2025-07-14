package context

import (
	"database/sql"

	"github.com/wellitonscheer/ticket-helper/internal/config"
	"github.com/wellitonscheer/ticket-helper/internal/milvus"
)

type AppContext struct {
	Config *config.Config
	Sqlite *sql.DB
	Milvus *milvus.MilvusClient
}
