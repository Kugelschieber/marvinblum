package tpl

import (
	"bytes"
	"fmt"
	"github.com/emvi/logbuch"
	"github.com/gosimple/slug"
	"html/template"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	templateDir = "template/*"
)

type Cache struct {
	tpl       *template.Template
	cache     map[string][]byte
	hotReload bool
	m         sync.RWMutex
}

func NewCache() *Cache {
	cache := &Cache{
		cache:     make(map[string][]byte),
		hotReload: os.Getenv("MB_HOT_RELOAD") == "true",
	}
	logbuch.Debug("Template cache hot reload", logbuch.Fields{"hot_reload": cache.hotReload})
	cache.load()
	return cache
}

func (cache *Cache) load() {
	logbuch.Debug("Loading templates")
	funcMap := template.FuncMap{
		"slug":     slug.Make,
		"format":   func(t time.Time, layout string) string { return t.Format(layout) },
		"multiply": func(f, x float64) float64 { return f * x },
		"round":    func(f float64) string { return fmt.Sprintf("%.2f", f) },
		"intRange": intRange,
	}
	var err error
	cache.tpl, err = template.New("").Funcs(funcMap).ParseGlob(templateDir)

	if err != nil {
		logbuch.Fatal("Error loading template", logbuch.Fields{"err": err})
	}

	logbuch.Debug("Templates loaded", logbuch.Fields{"hot_reload": cache.hotReload})
}

func (cache *Cache) Render(w http.ResponseWriter, name string, data interface{}) {
	cache.m.RLock()

	if cache.cache[name] == nil || cache.hotReload {
		cache.m.RUnlock()
		cache.m.Lock()
		defer cache.m.Unlock()
		logbuch.Debug("Rendering template", logbuch.Fields{"name": name})

		if cache.hotReload {
			logbuch.Debug("Reloading templates")
			cache.load()
		}

		var buffer bytes.Buffer

		if err := cache.tpl.ExecuteTemplate(&buffer, name, data); err != nil {
			logbuch.Error("Error executing template", logbuch.Fields{"err": err, "name": name})
			w.WriteHeader(http.StatusInternalServerError)
		}

		cache.cache[name] = buffer.Bytes()
	} else {
		cache.m.RUnlock()
	}

	if _, err := w.Write(cache.cache[name]); err != nil {
		logbuch.Error("Error sending response to client", logbuch.Fields{"err": err, "template": name})
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (cache *Cache) RenderWithoutCache(w http.ResponseWriter, name string, data interface{}) {
	if cache.hotReload {
		logbuch.Debug("Reloading templates")
		cache.load()
	}

	if err := cache.tpl.ExecuteTemplate(w, name, data); err != nil {
		logbuch.Error("Error executing template", logbuch.Fields{"err": err, "name": name})
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (cache *Cache) Clear() {
	cache.m.Lock()
	defer cache.m.Unlock()
	cache.cache = make(map[string][]byte)
}

func intRange(start, end int) []int {
	if end-start < 0 {
		return []int{}
	}

	r := make([]int, end-start)

	for i := 0; i < end-start; i++ {
		r[i] = start + i
	}

	return r
}
