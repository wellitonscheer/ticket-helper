package context

import (
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wellitonscheer/ticket-helper/internal/config"
)

type AppContext struct {
	Config *config.Config
	Sqlite *sql.DB
	PGVec  *pgxpool.Pool
}
