package tracking

import (
	"fmt"
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
	startTime := today()
	startTime = startTime.Add(-time.Hour * 24 * time.Duration(start-1))
	analyzer := pirsch.NewAnalyzer(store)
	visitors, err := analyzer.Visitors(&pirsch.Filter{From: startTime, To: today()})

	if err != nil {
		return "", ""
	}

	return getLabelsAndData(visitors)
}

func GetPageVisits(start int) []PageVisits {
	startTime := today()
	startTime = startTime.Add(-time.Hour * 24 * time.Duration(start-1))
	analyzer := pirsch.NewAnalyzer(store)
	visits, err := analyzer.PageVisits(&pirsch.Filter{From: startTime, To: today()})

	if err != nil {
		return nil
	}

	pageVisits := make([]PageVisits, len(visits))

	for i, visit := range visits {
		labels, data := getLabelsAndData(visit.Visits)
		pageVisits[i] = PageVisits{visit.Path, labels, data}
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
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
