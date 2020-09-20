package tracking

import (
	"context"
	"database/sql"
	"github.com/Kugelschieber/marvinblum.de/db"
	"github.com/emvi/logbuch"
	"github.com/emvi/pirsch"
	"os"
	"path/filepath"
)

const (
	geodbPath = "geodb"
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
		ReferrerDomainBlacklist:                   []string{"marvinblum.de"}, // I don't care about traffic from my own website
		ReferrerDomainBlacklistIncludesSubdomains: true,
		Sessions: true,
	})
	analyzer = pirsch.NewAnalyzer(store)
	processor := pirsch.NewProcessor(store)
	cancel := pirsch.RunAtMidnight(func() {
		processTrackingData(processor)
		updateGeoDB(tracker)
	})
	processTrackingData(processor)
	updateGeoDB(tracker)
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

func updateGeoDB(tracker *pirsch.Tracker) {
	licenseKey := os.Getenv("MB_GEOLITE2_LICENSE_KEY")

	if licenseKey == "" {
		return
	}

	if err := pirsch.GetGeoLite2(geodbPath, licenseKey); err != nil {
		logbuch.Error("Error loading GeoLite2", logbuch.Fields{"err": err})
		return
	}

	geodb, err := pirsch.NewGeoDB(filepath.Join(geodbPath, pirsch.GeoLite2Filename))

	if err != nil {
		logbuch.Error("Error creating GeoDB", logbuch.Fields{"err": err})
		return
	}

	tracker.SetGeoDB(geodb)
	logbuch.Info("GeoDB updated")
}
