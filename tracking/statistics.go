package tracking

import (
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/emvi/pirsch"
	"html/template"
	"sort"
	"strings"
	"time"
)

const (
	statisticsDateFormat = "2006-01-02"
)

type PageVisitors struct {
	Path     string
	Visitors int
	Labels   template.JS
	Data     template.JS
	Sessions template.JS
	Bounces  template.JS
}

func GetActiveVisitors() ([]pirsch.Stats, int) {
	visitors, total, err := analyzer.ActiveVisitors(nil, time.Minute*10)

	if err != nil {
		logbuch.Error("Error reading active visitors", logbuch.Fields{"err": err})
		return nil, 0
	}

	return visitors, total
}

func GetHourlyVisitorsToday() (template.JS, template.JS) {
	visitors, err := analyzer.VisitorHours(&pirsch.Filter{From: today(), To: today()})

	if err != nil {
		logbuch.Error("Error reading hourly visitors for today", logbuch.Fields{"err": err})
		return "", ""
	}

	return getLabelsAndDataHourly(visitors)
}

func GetTotalVisitors(startDate, endDate time.Time) (template.JS, template.JS, template.JS, template.JS) {
	visitors, err := analyzer.Visitors(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading visitor statistics", logbuch.Fields{"err": err})
		return "", "", "", ""
	}

	return getLabelsAndData(visitors)
}

func GetPageVisits(startDate, endDate time.Time) ([]PageVisitors, []PageVisitors) {
	visits, err := analyzer.PageVisitors(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading page statistics", logbuch.Fields{"err": err})
		return nil, nil
	}

	pageVisitors := make([]PageVisitors, len(visits))

	for i, visit := range visits {
		labels, data, sessions, bounces := getLabelsAndData(visit.Stats)
		pageVisitors[i] = PageVisitors{
			Path:     visit.Path,
			Visitors: sumVisitors(visit.Stats),
			Labels:   labels,
			Data:     data,
			Sessions: sessions,
			Bounces:  bounces,
		}
	}

	pageRank := make([]PageVisitors, len(pageVisitors))
	copy(pageRank, pageVisitors)
	sort.Slice(pageRank, func(i, j int) bool {
		return pageRank[i].Visitors > pageRank[j].Visitors
	})
	return pageVisitors, pageRank
}

func GetLanguages(startDate, endDate time.Time) []pirsch.LanguageStats {
	languages, err := analyzer.Languages(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading language statistics", logbuch.Fields{"err": err})
		return nil
	}

	if len(languages) > 10 {
		return languages[:10]
	}

	return languages
}

func GetReferrer(startDate, endDate time.Time) []pirsch.ReferrerStats {
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

func GetOS(startDate, endDate time.Time) []pirsch.OSStats {
	os, err := analyzer.OS(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading OS statistics", logbuch.Fields{"err": err})
		return nil
	}

	return os
}

func GetBrowser(startDate, endDate time.Time) []pirsch.BrowserStats {
	browser, err := analyzer.Browser(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading browser statistics", logbuch.Fields{"err": err})
		return nil
	}

	return browser
}

func GetCountry(startDate, endDate time.Time) []pirsch.CountryStats {
	countries, err := analyzer.Country(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading country statistics", logbuch.Fields{"err": err})
		return nil
	}

	for i := range countries {
		countries[i].CountryCode.String = strings.ToUpper(countries[i].CountryCode.String)
	}

	return countries
}

func GetPlatform(startDate, endDate time.Time) *pirsch.VisitorStats {
	return analyzer.Platform(&pirsch.Filter{From: startDate, To: endDate})
}

func GetVisitorTimeOfDay(startDate, endDate time.Time) []pirsch.TimeOfDayVisitors {
	visitors, err := analyzer.TimeOfDay(&pirsch.Filter{From: startDate, To: endDate})

	if err != nil {
		logbuch.Error("Error reading visitor time of day statistics", logbuch.Fields{"err": err})
		return nil
	}

	return visitors
}

func sumVisitors(stats []pirsch.Stats) int {
	sum := 0

	for _, s := range stats {
		sum += s.Visitors
	}

	return sum
}

func getLabelsAndData(visitors []pirsch.Stats) (template.JS, template.JS, template.JS, template.JS) {
	var labels strings.Builder
	var dp strings.Builder
	var sessions strings.Builder
	var bounces strings.Builder

	for _, point := range visitors {
		labels.WriteString(fmt.Sprintf("'%s',", point.Day.Format(statisticsDateFormat)))
		dp.WriteString(fmt.Sprintf("%d,", point.Visitors))
		sessions.WriteString(fmt.Sprintf("%d,", point.Sessions))
		bounces.WriteString(fmt.Sprintf("%d,", point.Bounces))
	}

	labelsStr := labels.String()
	dataStr := dp.String()
	sessionsStr := sessions.String()
	bouncesStr := bounces.String()
	return template.JS(labelsStr[:len(labelsStr)-1]),
		template.JS(dataStr[:len(dataStr)-1]),
		template.JS(sessionsStr[:len(sessionsStr)-1]),
		template.JS(bouncesStr[:len(bouncesStr)-1])
}

func getLabelsAndDataHourly(visitors []pirsch.VisitorTimeStats) (template.JS, template.JS) {
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
