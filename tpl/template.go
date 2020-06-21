package tpl

import (
	"bytes"
	"github.com/emvi/logbuch"
	"github.com/emvi/pirsch"
	"github.com/gosimple/slug"
	"html/template"
	"net/http"
	"os"
	"time"
)

const (
	templateDir = "template/*"
)

var (
	tpl       *template.Template
	tplCache  = make(map[string][]byte)
	hotReload bool
)

var funcMap = template.FuncMap{
	"slug": slug.Make,
	"format": func(t time.Time, layout string) string {
		return t.Format(layout)
	},
}

func LoadTemplate() {
	logbuch.Debug("Loading templates")
	var err error
	tpl, err = template.New("").Funcs(funcMap).ParseGlob(templateDir)

	if err != nil {
		logbuch.Fatal("Error loading template", logbuch.Fields{"err": err})
	}

	hotReload = os.Getenv("MB_HOT_RELOAD") == "true"
	logbuch.Debug("Templates loaded", logbuch.Fields{"hot_reload": hotReload})
}

func renderTemplate(name string) {
	logbuch.Debug("Rendering template", logbuch.Fields{"name": name})
	var buffer bytes.Buffer

	if err := tpl.ExecuteTemplate(&buffer, name, nil); err != nil {
		logbuch.Fatal("Error executing template", logbuch.Fields{"err": err, "name": name})
	}

	tplCache[name] = buffer.Bytes()
}

func Get() *template.Template {
	return tpl
}

func ServeTemplate(name string, tracker *pirsch.Tracker) http.HandlerFunc {
	// render once so we have it in cache
	renderTemplate(name)

	return func(w http.ResponseWriter, r *http.Request) {
		tracker.Hit(r)

		if hotReload {
			LoadTemplate()
			renderTemplate(name)
		}

		if _, err := w.Write(tplCache[name]); err != nil {
			logbuch.Error("Error returning page to client", logbuch.Fields{"err": err, "name": name})
		}
	}
}
