package tools

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type database struct{}

func NewDatabase() *database {
	return &database{}
}

func (d *database) GetGormConnection(host, port, user, password, db, timezone string) (*gorm.DB, error) {
	return gorm.Open(postgres.New(postgres.Config{
		DSN: fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
			host,
			user,
			password,
			db,
			port,
			timezone,
		), // data source name, refer https://github.com/jackc/pgx
		PreferSimpleProtocol: true, // disables implicit prepared statement usage. By default pgx automatically uses the extended protocol
	}), &gorm.Config{})
}
