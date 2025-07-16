package pgvec

import (
	"context"
	"fmt"
	"os"

	appContext "github.com/wellitonscheer/ticket-helper/internal/context"
)

func InitiatePGVec(ctx *appContext.AppContext) {
	fileMigration, err := os.ReadFile("./internal/database/pgvec/migrations/1_create_tables.sql")
	if err != nil {
		fmt.Printf("error to read migration file: %v\n", err)
	}

	_, err = ctx.PGVec.Exec(context.Background(), string(fileMigration))
	if err != nil {
		fmt.Printf("error to execute migration: %v\n", err)
	}
}
