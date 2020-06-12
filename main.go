package main

import (
	"bytes"
	"github.com/NYTimes/gziphandler"
	"github.com/caddyserver/certmagic"
	"github.com/emvi/logbuch"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"html/template"
	"net/http"
	"os"
	"strings"
)

const (
	staticDir       = "static"
	staticDirPrefix = "/static/"
	templateDir     = "template/*"
	logTimeFormat   = "2006-01-02_15:04:05"
	envPrefix       = "MB_"
)

var (
	tpl       *template.Template
	tplCache  = make(map[string][]byte)
	hotReload bool
)

func configureLog() {
	logbuch.SetFormatter(logbuch.NewFieldFormatter(logTimeFormat, "\t\t"))
	logbuch.Info("Configure logging...")
	level := strings.ToLower(os.Getenv("MB_LOGLEVEL"))

	if level == "debug" {
		logbuch.SetLevel(logbuch.LevelDebug)
	} else if level == "info" {
		logbuch.SetLevel(logbuch.LevelInfo)
	} else {
		logbuch.SetLevel(logbuch.LevelWarning)
	}
}

func logEnvConfig() {
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, envPrefix) {
			pair := strings.Split(e, "=")
			logbuch.Info(pair[0] + "=" + pair[1])
		}
	}
}

func loadTemplate() {
	logbuch.Debug("Loading templates")
	var err error
	tpl, err = template.ParseGlob(templateDir)

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

func serveTemplate(name string) http.HandlerFunc {
	// render once so we have it in cache
	renderTemplate(name)

	return func(w http.ResponseWriter, r *http.Request) {
		if hotReload {
			loadTemplate()
			renderTemplate(name)
		}

		if _, err := w.Write(tplCache[name]); err != nil {
			logbuch.Error("Error returning page to client", logbuch.Fields{"err": err, "name": name})
		}
	}
}

func setupRouter() *mux.Router {
	router := mux.NewRouter()
	router.PathPrefix(staticDirPrefix).Handler(http.StripPrefix(staticDirPrefix, gziphandler.GzipHandler(http.FileServer(http.Dir(staticDir)))))
	router.Handle("/legal", serveTemplate("legal.html"))
	router.Handle("/blog", serveTemplate("blog.html"))
	router.Handle("/", serveTemplate("about.html"))
	router.NotFoundHandler = serveTemplate("notfound.html")
	return router
}

func configureCors(router *mux.Router) http.Handler {
	logbuch.Info("Configuring CORS...")

	origins := strings.Split(os.Getenv("MB_ALLOWED_ORIGINS"), ",")
	c := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            strings.ToLower(os.Getenv("MB_CORS_LOGLEVEL")) == "debug",
	})
	return c.Handler(router)
}

func start(handler http.Handler) {
	logbuch.Info("Starting server...")

	if strings.ToLower(os.Getenv("MB_TLS")) == "true" {
		logbuch.Info("TLS enabled")
		certmagic.DefaultACME.Agreed = true
		certmagic.DefaultACME.Email = os.Getenv("MB_TLS_EMAIL")
		certmagic.DefaultACME.CA = certmagic.LetsEncryptProductionCA

		if err := certmagic.HTTPS(strings.Split(os.Getenv("MB_DOMAIN"), ","), handler); err != nil {
			logbuch.Fatal("Error starting server", logbuch.Fields{"err": err})
		}
	} else {
		if err := http.ListenAndServe(os.Getenv("MB_HOST"), handler); err != nil {
			logbuch.Fatal("Error starting server", logbuch.Fields{"err": err})
		}
	}
}

func main() {
	configureLog()
	logEnvConfig()
	loadTemplate()
	router := setupRouter()
	corsConfig := configureCors(router)
	start(corsConfig)
}
