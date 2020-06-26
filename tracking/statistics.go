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

func GetTotalVisitors(startDate, endDate time.Time) (template.JS, template.JS) {
	visitors, err := analyzer.Visitors(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading visitor statistics", logbuch.Fields{"err": err})
		return "", ""
	}

	return getLabelsAndData(visitors)
}

func GetPageVisits(startDate, endDate time.Time) []PageVisits {
	visits, err := analyzer.PageVisits(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading page statistics", logbuch.Fields{"err": err})
		return nil
	}

	pageVisits := make([]PageVisits, len(visits))

	for i, visit := range visits {
		labels, data := getLabelsAndData(visit.Visits)
		pageVisits[i] = PageVisits{visit.Path, labels, data}
	}

	return pageVisits
}

func GetLanguages(startDate, endDate time.Time) []pirsch.VisitorLanguage {
	languages, _, err := analyzer.Languages(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading language statistics", logbuch.Fields{"err": err})
		return nil
	}

	if len(languages) > 10 {
		return languages[:10]
	}

	return languages
}

func GetHourlyVisitors(startDate, endDate time.Time) (template.JS, template.JS) {
	visitors, err := analyzer.HourlyVisitors(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading hourly visitors", logbuch.Fields{"err": err})
		return "", ""
	}

	return getLabelsAndDataHourly(visitors)
}

func GetHourlyVisitorsToday() (template.JS, template.JS) {
	visitors, err := analyzer.HourlyVisitors(&pirsch.Filter{From: today(), To: today()})

	if err != nil {
		logbuch.Error("Error reading hourly visitors for today", logbuch.Fields{"err": err})
		return "", ""
	}

	return getLabelsAndDataHourly(visitors)
}

func GetActiveVisitors() int {
	visitors, err := analyzer.ActiveVisitors(time.Minute * 5)

	if err != nil {
		logbuch.Error("Error reading active visitors", logbuch.Fields{"err": err})
		return 0
	}

	return visitors
}

func getStartTime(start int) time.Time {
	startTime := today()
	return startTime.Add(-time.Hour * 24 * time.Duration(start-1))
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

func getLabelsAndDataHourly(visitors []pirsch.HourlyVisitors) (template.JS, template.JS) {
	var labels strings.Builder
	var dp strings.Builder

	for _, point := range visitors {
		labels.WriteString(fmt.Sprintf("'%d',", point.Hour))
		dp.WriteString(fmt.Sprintf("%d,", point.Visitors))
	}

	labelsStr := labels.String()
	dataStr := dp.String()
	return template.JS(labelsStr[:len(labelsStr)-1]), template.JS(dataStr[:len(dataStr)-1])
}

func today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
