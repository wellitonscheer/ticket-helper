package osticket

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/wellitonscheer/ticket-helper/internal/database/osticket/tickmodel"
)

func StartChecker(db *sqlx.DB) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	checkUpdatedTicket(db)

	for range ticker.C {
		checkUpdatedTicket(db)
	}
}

func checkUpdatedTicket(db *sqlx.DB) {
	ticket := tickmodel.OstTicket{}
	err := db.Get(&ticket, "SELECT * FROM ost_ticket WHERE ticket_id=?", 33683)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	fmt.Printf("%#v\n", ticket)
}
