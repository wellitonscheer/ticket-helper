package pgvec

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxvec "github.com/pgvector/pgvector-go/pgx"
	"github.com/wellitonscheer/ticket-helper/internal/config"
)

func NewPGVectorConnection(pgVecConf config.PGVectorConfig) *pgxpool.Pool {
	fmt.Println("Connecting to PGVector now.")

	connString := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s", pgVecConf.PostgresUser, pgVecConf.PostgresPassword, pgVecConf.PostgresPort, pgVecConf.PostgresDB)

	connConf, err := pgxpool.ParseConfig(connString)
	if err != nil {
		fmt.Printf("failed to create pgxpoll config\n")
		panic(err)
	}

	connConf.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvec.RegisterTypes(ctx, conn)
	}

	pgpool, err := pgxpool.NewWithConfig(context.Background(), connConf)
	if err != nil {
		fmt.Printf("unable to create connection pool (connConf=%+v)", connConf)
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
