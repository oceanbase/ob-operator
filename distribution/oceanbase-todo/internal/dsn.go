package internal

import (
	"errors"
	"os"
	"strings"
)

type DSN struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func (d DSN) String() string {
	return d.User + ":" + d.Password + "@tcp(" + d.Host + ":" + d.Port + ")/" + d.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
}

func NewDSN(prefixes ...string) (*DSN, error) {
	prefix := ""
	if len(prefixes) > 0 {
		prefix = strings.Join(prefixes, "_") + "_"
	}
	dsn := DSN{
		Host:     os.Getenv(prefix + "DB_HOST"),
		Port:     os.Getenv(prefix + "DB_PORT"),
		User:     os.Getenv(prefix + "DB_USER"),
		Password: os.Getenv(prefix + "DB_PASSWORD"),
		Database: os.Getenv(prefix + "DB_DATABASE"),
	}

	if dsn.Host == "" || dsn.User == "" || dsn.Database == "" || dsn.Port == "" {
		return nil, errors.New("missing required environment variables. Check DB_HOST, DB_PORT, DB_USER and DB_DATABASE please.")
	}
	return &dsn, nil
}
