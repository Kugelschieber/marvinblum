package main

import (
	"github.com/Kugelschieber/marvinblum.de/blog"
	"github.com/Kugelschieber/marvinblum.de/tpl"
	"github.com/Kugelschieber/marvinblum.de/tracking"
	"github.com/NYTimes/gziphandler"
	"github.com/caddyserver/certmagic"
	emvi "github.com/emvi/api-go"
	"github.com/emvi/logbuch"
	"github.com/emvi/pirsch"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	staticDir       = "static"
	staticDirPrefix = "/static/"
	logTimeFormat   = "2006-01-02_15:04:05"
	envPrefix       = "MB_"
)

var (
	tracker      *pirsch.Tracker
	tplCache     *tpl.Cache
	blogInstance *blog.Blog
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

func serveAbout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracker.Hit(r)
		tplCache.Render(w, "about.html", struct {
			Articles []emvi.Article
		}{
			blogInstance.GetLatestArticles(),
		})
	}
}

func serveLegal() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracker.Hit(r)
		tplCache.Render(w, "legal.html", nil)
	}
}

func serveBlogPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracker.Hit(r)
		tplCache.Render(w, "blog.html", struct {
			Articles map[int][]emvi.Article
		}{
			blogInstance.GetArticles(),
		})
	}
}

func serveBlogArticle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracker.Hit(r)
		vars := mux.Vars(r)
		slug := strings.Split(vars["slug"], "-")

		if len(slug) == 0 {
			http.Redirect(w, r, "/notfound", http.StatusFound)
			return
		}

		article := blogInstance.GetArticle(slug[len(slug)-1])

		if len(article.Id) == 0 {
			http.Redirect(w, r, "/notfound", http.StatusFound)
			return
		}

		tplCache.Render(w, "article.html", struct {
			Title     string
			Content   template.HTML
			Published time.Time
		}{
			article.LatestArticleContent.Title,
			template.HTML(article.LatestArticleContent.Content),
			article.Published,
		})
	}
}

func serveNotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tplCache.Render(w, "notfound.html", nil)
	}
}

func setupRouter() *mux.Router {
	router := mux.NewRouter()
	router.PathPrefix(staticDirPrefix).Handler(http.StripPrefix(staticDirPrefix, gziphandler.GzipHandler(http.FileServer(http.Dir(staticDir)))))
	router.Handle("/blog/{slug}", serveBlogArticle())
	router.Handle("/blog", serveBlogPage())
	router.Handle("/legal", serveLegal())
	router.Handle("/", serveAbout())
	router.NotFoundHandler = serveNotFound()
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
	tracker = tracking.NewTracker()
	tplCache = tpl.NewCache()
	blogInstance = blog.NewBlog(tplCache)
	router := setupRouter()
	corsConfig := configureCors(router)
	start(corsConfig)
}
