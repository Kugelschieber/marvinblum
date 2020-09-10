package tracking

import (
	"context"
	"database/sql"
	"github.com/Kugelschieber/marvinblum.de/db"
	"github.com/emvi/logbuch"
	"github.com/emvi/pirsch"
	"os"
)

const (
	domain = "marvinblum.de"
)

var (
	store    pirsch.Store
	analyzer *pirsch.Analyzer
)

func NewTracker() (*pirsch.Tracker, context.CancelFunc) {
	logbuch.Info("Connecting to database...")
	conn, err := sql.Open("postgres", db.GetConnectionString())

	if err != nil {
		logbuch.Fatal("Error connecting to database", logbuch.Fields{"err": err})
		return nil, nil
	}

	if err := conn.Ping(); err != nil {
		logbuch.Fatal("Error pinging database", logbuch.Fields{"err": err})
		return nil, nil
	}

	store = pirsch.NewPostgresStore(conn, nil)
	tracker := pirsch.NewTracker(store, os.Getenv("MB_TRACKING_SALT"), &pirsch.TrackerConfig{
		// I don't care about traffic from my own website
		ReferrerDomainBlacklist:                   []string{domain},
		ReferrerDomainBlacklistIncludesSubdomains: true,
	})
	analyzer = pirsch.NewAnalyzer(store)
	processor := pirsch.NewProcessor(store)
	cancel := pirsch.RunAtMidnight(func() {
		processTrackingData(processor)
	})
	processTrackingData(processor) // run on startup
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
