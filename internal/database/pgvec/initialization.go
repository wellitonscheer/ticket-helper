package pgvec

import (
	"context"
	"fmt"
	"os"

	appContext "github.com/wellitonscheer/ticket-helper/internal/context"
)

const migrationFolder string = "./internal/database/pgvec/migrations"

func InitiatePGVec(ctx *appContext.AppContext) {
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

		_, err = ctx.PGVec.Exec(context.Background(), string(fileMigration))
		if err != nil {
			fmt.Println("error to execute migration")
			panic(err)
		}
	}

	fmt.Println("\nPGVector migrations applied\n")
}
