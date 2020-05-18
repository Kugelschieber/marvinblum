package main

import (
	"github.com/caddyserver/certmagic"
	"github.com/emvi/logbuch"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"net/http"
	"os"
	"strings"
)

const (
	staticDir       = "static"
	staticDirPrefix = "/"
	logTimeFormat   = "2006-01-02_15:04:05"
	envPrefix       = "MB_"
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

func setupRouter() *mux.Router {
	router := mux.NewRouter()
	router.PathPrefix(staticDirPrefix).Handler(http.StripPrefix(staticDirPrefix, http.FileServer(http.Dir(staticDir))))
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
	router := setupRouter()
	corsConfig := configureCors(router)
	start(corsConfig)
}
