package db

import (
	"fmt"
	"github.com/emvi/logbuch"
	"os"
	"strconv"
	"time"
)

const (
	connectionString = `host=%s port=%s user=%s password=%s dbname=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s connectTimeout=%s timezone=%s`
)

func GetConnectionString() string {
	host := os.Getenv("MB_DB_HOST")
	port := os.Getenv("MB_DB_PORT")
	user := os.Getenv("MB_DB_USER")
	password := os.Getenv("MB_DB_PASSWORD")
	schema := os.Getenv("MB_DB_SCHEMA")
	sslMode := os.Getenv("MB_DB_SSLMODE")
	sslCert := os.Getenv("MB_DB_SSLCERT")
	sslKey := os.Getenv("MB_DB_SSLKEY")
	sslRootCert := os.Getenv("MB_DB_SSLROOTCERT")
	zone, offset := time.Now().Zone()
	timezone := zone + strconv.Itoa(-offset/3600)
	logbuch.Info("Setting time zone", logbuch.Fields{"timezone": timezone})
	return fmt.Sprintf(connectionString, host, port, user, password, schema, sslMode, sslCert, sslKey, sslRootCert, "30", timezone)
}
