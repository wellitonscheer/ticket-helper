package pgvec

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wellitonscheer/ticket-helper/internal/config"
)

func NewPGVectorConnection(pgVecConf *config.PGVectorConfig) *pgxpool.Pool {
	fmt.Println("Connecting to PGVector now.")

	connString := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s", pgVecConf.PostgresUser, pgVecConf.PostgresPassword, pgVecConf.PostgresPort, pgVecConf.PostgresDB)

	pgpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		fmt.Printf("Unable to create connection pool (connString=%s)", connString)
		panic(err)
	}

	var greeting string
	err = pgpool.QueryRow(context.Background(), "select 'Hello from pgvector!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed")
		panic(err)
	}

	fmt.Println(greeting)

	fmt.Println("PGVector connected.")

	return pgpool
}
