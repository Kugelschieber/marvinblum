package tracking

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/emvi/pirsch"
	"os"
	"strconv"
	"time"
)

const (
	connectionString = `host=%s port=%s user=%s password=%s dbname=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s connectTimeout=%s timezone=%s`
	domain           = "marvinblum.de"
)

var (
	store    pirsch.Store
	analyzer *pirsch.Analyzer
)

func NewTracker() (*pirsch.Tracker, context.CancelFunc) {
	logbuch.Info("Connecting to database...")
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
	connectionStr := fmt.Sprintf(connectionString, host, port, user, password, schema, sslMode, sslCert, sslKey, sslRootCert, "30", timezone)
	conn, err := sql.Open("postgres", connectionStr)

	if err != nil {
		logbuch.Fatal("Error connecting to database", logbuch.Fields{"err": err})
		return nil, nil
	}

	if err := conn.Ping(); err != nil {
		logbuch.Fatal("Error pinging database", logbuch.Fields{"err": err})
		return nil, nil
	}

	store = pirsch.NewPostgresStore(conn)
	tracker := pirsch.NewTracker(store, os.Getenv("MB_TRACKING_SALT"), &pirsch.TrackerConfig{
		// I don't care about traffic from my own website
		RefererDomainBlacklist:                   []string{domain},
		RefererDomainBlacklistIncludesSubdomains: true,
	})
	analyzer = pirsch.NewAnalyzer(store)
	processor := pirsch.NewProcessor(store, nil)
	cancel := pirsch.RunAtMidnight(func() {
		processTrackingData(processor)
	})
	return tracker, cancel
}

func processTrackingData(processor *pirsch.Processor) {
	logbuch.Info("Processing tracking data...")

	defer func() {
		if err := recover(); err != nil {
			logbuch.Error("Error processing tracking data", logbuch.Fields{"err": err})
		}
	}()

	if err := processor.Process(); err != nil {
		logbuch.Error("Error processing tracking data", logbuch.Fields{"err": err})
	} else {
		logbuch.Info("Done processing tracking data")
	}
}
