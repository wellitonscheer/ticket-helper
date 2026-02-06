package osticket

import (
	"fmt"
	"log"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/wellitonscheer/ticket-helper/internal/config"
)

func NewOsTicketConnection(osTickConf config.OsTicketConfig) *sqlx.DB {
	fmt.Println("Connecting to OsTicket now.")
	cfg := mysql.NewConfig()
	cfg.Addr = fmt.Sprintf("%s:%s", osTickConf.Host, osTickConf.Port)
	cfg.DBName = osTickConf.DB
	cfg.User = osTickConf.User
	cfg.Passwd = osTickConf.Password
	cfg.Net = "tcp"
	cfg.ParseTime = true

	db, err := sqlx.Connect("mysql", cfg.FormatDSN())
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
