package main

import (
	"github.com/Kugelschieber/marvinblum.de/blog"
	"github.com/Kugelschieber/marvinblum.de/tpl"
	"github.com/NYTimes/gziphandler"
	"github.com/caddyserver/certmagic"
	emvi "github.com/emvi/api-go"
	"github.com/emvi/logbuch"
	"github.com/gorilla/mux"
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
		data := struct {
			Articles []emvi.Article
		}{
			blogInstance.GetLatestArticles(),
		}

		if err := tpl.Get().ExecuteTemplate(w, "about.html", data); err != nil {
			logbuch.Error("Error executing blog template", logbuch.Fields{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func serveBlogPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Articles map[int][]emvi.Article
		}{
			blogInstance.GetArticles(),
		}

		if err := tpl.Get().ExecuteTemplate(w, "blog.html", data); err != nil {
			logbuch.Error("Error executing blog template", logbuch.Fields{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
		}
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

		data := struct {
			Title     string
			Content   template.HTML
			Published time.Time
		}{
			article.LatestArticleContent.Title,
			template.HTML(article.LatestArticleContent.Content),
			article.Published,
		}

		if err := tpl.Get().ExecuteTemplate(w, "article.html", data); err != nil {
			logbuch.Error("Error executing blog article template", logbuch.Fields{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func setupRouter() *mux.Router {
	router := mux.NewRouter()
	router.PathPrefix(staticDirPrefix).Handler(http.StripPrefix(staticDirPrefix, gziphandler.GzipHandler(http.FileServer(http.Dir(staticDir)))))
	router.Handle("/blog/{slug}", serveBlogArticle())
	router.Handle("/blog", serveBlogPage())
	router.Handle("/legal", tpl.ServeTemplate("legal.html"))
	router.Handle("/", serveAbout())
	router.NotFoundHandler = tpl.ServeTemplate("notfound.html")
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
	tpl.LoadTemplate()
	blogInstance = blog.NewBlog()
	router := setupRouter()
	corsConfig := configureCors(router)
	start(corsConfig)
}
