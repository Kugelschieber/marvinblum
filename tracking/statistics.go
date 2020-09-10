package tracking

import (
	"html/template"
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

/*func GetTotalVisitors(startDate, endDate time.Time) (template.JS, template.JS) {
	visitors, err := analyzer.Visitors(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading visitor statistics", logbuch.Fields{"err": err})
		return "", ""
	}

	return getLabelsAndData(visitors)
}

func GetPageVisits(startDate, endDate time.Time) []PageVisits {
	visits, err := analyzer.PageVisitors(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading page statistics", logbuch.Fields{"err": err})
		return nil
	}

	pageVisits := make([]PageVisits, len(visits))

	for i, visit := range visits {
		labels, data := getLabelsAndData(visit.VisitorsPerDay)
		pageVisits[i] = PageVisits{visit.Path.String, labels, data}
	}

	return pageVisits
}

func GetPages(startDate, endDate time.Time) []pirsch.Stats {
	pages, err := analyzer.Pages(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading page statistics", logbuch.Fields{"err": err})
		return nil
	}

	if len(pages) > 10 {
		return pages[:10]
	}

	return pages
}

func GetLanguages(startDate, endDate time.Time) []pirsch.Stats {
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

func GetReferrer(startDate, endDate time.Time) []pirsch.Stats {
	referrer, err := analyzer.Referrer(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading referrer statistics", logbuch.Fields{"err": err})
		return nil
	}

	if len(referrer) > 10 {
		return referrer[:10]
	}

	return referrer
}

func GetOS(startDate, endDate time.Time) []pirsch.Stats {
	os, err := analyzer.OS(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading OS statistics", logbuch.Fields{"err": err})
		return nil
	}

	return os
}

func GetBrowser(startDate, endDate time.Time) []pirsch.Stats {
	browser, err := analyzer.Browser(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading browser statistics", logbuch.Fields{"err": err})
		return nil
	}

	return browser
}

func GetPlatform(startDate, endDate time.Time) *pirsch.Stats {
	platform, err := analyzer.Platform(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading platform statistics", logbuch.Fields{"err": err})
		return nil
	}

	return platform
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
	visitors, err := analyzer.ActiveVisitors(pirsch.NullTenant, time.Minute*5)

	if err != nil {
		logbuch.Error("Error reading active visitors", logbuch.Fields{"err": err})
		return 0
	}

	return visitors
}

func GetActiveVisitorPages() []pirsch.Stats {
	pages, err := analyzer.ActiveVisitorsPages(pirsch.NullTenant, time.Second*30)

	if err != nil {
		logbuch.Error("Error reading active visitor pages", logbuch.Fields{"err": err})
		return nil
	}

	return pages
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

func getLabelsAndDataHourly(visitors []pirsch.Stats) (template.JS, template.JS) {
	var labels strings.Builder
	var dp strings.Builder

	for _, point := range visitors {
		labels.WriteString(fmt.Sprintf("'%d',", point.Hour))
		dp.WriteString(fmt.Sprintf("%d,", point.Visitors))
	}

	labelsStr := labels.String()
	dataStr := dp.String()
	return template.JS(labelsStr[:len(labelsStr)-1]), template.JS(dataStr[:len(dataStr)-1])
}*/

func today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
