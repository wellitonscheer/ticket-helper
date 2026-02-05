package osticket

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/go-sql-driver/mysql"
	"github.com/wellitonscheer/ticket-helper/internal/config"
)

func NewOsTicketConnection(osTickConf config.OsTicketConfig) *sql.DB {
	fmt.Println("Connecting to OsTicket now.")
	cfg := mysql.NewConfig()
	cfg.Addr = fmt.Sprintf("%s:%s", osTickConf.Host, osTickConf.Port)
	cfg.DBName = osTickConf.DB
	cfg.User = osTickConf.User
	cfg.Passwd = osTickConf.Password
	cfg.Net = "tcp"

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("OsTicket Connected!")

	return db
}
