package tracking

import (
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/emvi/pirsch"
	"html/template"
	"strings"
	"time"
)

const (
	statisticsDateFormat = "2006-01-02"
)

type PageVisits struct {
	Path   string
	Labels template.JS
	Data   template.JS
}

func GetTotalVisitors(start int) (template.JS, template.JS) {
	query := `SELECT "date" "day",
		CASE WHEN "visitors_per_day".visitors IS NULL THEN 0 ELSE "visitors_per_day".visitors END
		FROM (SELECT * FROM generate_series(date($1), date(now()), interval '1 day') "date") AS date_series
		LEFT JOIN "visitors_per_day" ON date("visitors_per_day"."day") = date("date")
		ORDER BY "date" ASC`
	startTime := today()
	startTime = startTime.Add(-time.Hour * 24 * time.Duration(start-1))
	logbuch.Debug("Reading total visitors since", logbuch.Fields{"since": startTime})
	var visitors []pirsch.VisitorsPerDay

	if err := db.Select(&visitors, query, startTime); err != nil {
		logbuch.Error("Error reading total visitors", logbuch.Fields{"err": err, "since": startTime})
		return "", ""
	}

	today := today()
	visitorsToday, err := store.VisitorsPerDay(today)

	if err != nil {
		logbuch.Error("Error reading total visitors for today", logbuch.Fields{"err": err, "since": startTime})
		return "", ""
	}

	visitors[len(visitors)-1].Visitors = visitorsToday
	return getLabelsAndData(visitors)
}

func GetPageVisits(start int) []PageVisits {
	pathsQuery := `SELECT * FROM (SELECT DISTINCT "path" FROM "visitors_per_page" WHERE day >= $1) AS paths ORDER BY length("path") ASC`
	query := `SELECT "date" "day",
		CASE WHEN "visitors_per_page".visitors IS NULL THEN 0 ELSE "visitors_per_page".visitors END
		FROM (SELECT * FROM generate_series(date($1), date(now() - INTERVAL '1 day'), interval '1 day') "date") AS date_series
		LEFT JOIN "visitors_per_page" ON date("visitors_per_page"."day") = date("date") AND "visitors_per_page"."path" = $2
		ORDER BY "date" ASC, length("visitors_per_page"."path") ASC`
	startTime := today()
	startTime = startTime.Add(-time.Hour * 24 * time.Duration(start-1))
	logbuch.Debug("Reading page visits", logbuch.Fields{"since": startTime})
	var paths []string

	if err := db.Select(&paths, pathsQuery, startTime); err != nil {
		logbuch.Error("Error reading distinct paths", logbuch.Fields{"err": err})
		return nil
	}

	// TODO add visitors for today
	pageVisits := make([]PageVisits, len(paths))

	for i, path := range paths {
		var visitors []pirsch.VisitorsPerDay

		if err := db.Select(&visitors, query, startTime, path); err != nil {
			logbuch.Error("Error reading visitor for path", logbuch.Fields{"err": err, "since": startTime})
			return nil
		}

		labels, data := getLabelsAndData(visitors)
		pageVisits[i] = PageVisits{path, labels, data}
	}

	return pageVisits
}

func getLabelsAndData(visitors []pirsch.VisitorsPerDay) (template.JS, template.JS) {
	var labels strings.Builder
	var dp strings.Builder

	for _, point := range visitors {
		labels.WriteString(fmt.Sprintf("'%s',", point.Day.Format(statisticsDateFormat)))
		dp.WriteString(fmt.Sprintf("%d,", point.Visitors))
	}

	labelsStr := labels.String()
	dataStr := dp.String()
	return template.JS(labelsStr[:len(labelsStr)-1]), template.JS(dataStr[:len(dataStr)-1])
}

func today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
