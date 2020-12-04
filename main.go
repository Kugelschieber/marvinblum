package main

import (
	"context"
	"github.com/Kugelschieber/marvinblum/blog"
	"github.com/Kugelschieber/marvinblum/tpl"
	"github.com/NYTimes/gziphandler"
	emvi "github.com/emvi/api-go"
	"github.com/emvi/logbuch"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/pirsch-analytics/pirsch-go-sdk"
	"github.com/rs/cors"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	staticDir       = "static"
	staticDirPrefix = "/static/"
	logTimeFormat   = "2006-01-02_15:04:05"
	envPrefix       = "MB_"
	shutdownTimeout = time.Second * 30
)

var (
	client       *pirsch.Client
	tplCache     *tpl.Cache
	blogInstance *blog.Blog
)

func serveAbout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		go hit(r)
		tplCache.Render(w, "about.html", struct {
			Articles []emvi.Article
		}{
			blogInstance.GetLatestArticles(),
		})
	}
}

func serveLegal() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		go hit(r)
		tplCache.Render(w, "legal.html", nil)
	}
}

func serveBlogPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		go hit(r)
		tplCache.Render(w, "blog.html", struct {
			Articles map[int][]emvi.Article
		}{
			blogInstance.GetArticles(),
		})
	}
}

func serveBlogArticle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		// track the hit if the article was found, otherwise we don't care
		go hit(r)

		tplCache.RenderWithoutCache(w, "article.html", struct {
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

func serveTracking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://marvinblum.pirsch.io/", http.StatusFound)
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
	router.Handle("/tracking", serveTracking())
	router.Handle("/", serveAbout())
	router.NotFoundHandler = serveNotFound()
	return router
}

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
	var server http.Server
	server.Handler = handler
	server.Addr = os.Getenv("MB_HOST")

	go func() {
		sigint := make(chan os.Signal)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		logbuch.Info("Shutting down server...")
		ctx, _ := context.WithTimeout(context.Background(), shutdownTimeout)

		if err := server.Shutdown(ctx); err != nil {
			logbuch.Fatal("Error shutting down server gracefully", logbuch.Fields{"err": err})
		}
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logbuch.Fatal("Error starting server", logbuch.Fields{"err": err})
	}

	logbuch.Info("Server shut down")
}

func hit(r *http.Request) {
	if err := client.Hit(r); err != nil {
		logbuch.Warn("Error sending page hit to pirsch", logbuch.Fields{"err": err})
	}
}

func main() {
	configureLog()
	logEnvConfig()
	client = pirsch.NewClient(os.Getenv("MB_PIRSCH_CLIENT_ID"),
		os.Getenv("MB_PIRSCH_CLIENT_SECRET"),
		os.Getenv("MB_PIRSCH_HOSTNAME"),
		nil)
	tplCache = tpl.NewCache()
	blogInstance = blog.NewBlog(tplCache)
	router := setupRouter()
	corsConfig := configureCors(router)
	start(corsConfig)
}
