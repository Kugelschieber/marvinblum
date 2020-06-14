package blog

import (
	"github.com/Kugelschieber/marvinblum.de/tpl"
	emvi "github.com/emvi/api-go"
	"github.com/emvi/logbuch"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	blogCacheTime     = time.Hour
	maxLatestArticles = 3
)

var (
	blog Blog
)

type Blog struct {
	client       *emvi.Client
	articles     map[string]emvi.Article // id -> article
	articlesYear map[int][]emvi.Article  // year -> articles
	nextUpdate   time.Time
}

func InitBlog() {
	logbuch.Info("Initializing blog")
	blog.client = emvi.NewClient(os.Getenv("MB_EMVI_CLIENT_ID"),
		os.Getenv("MB_EMVI_CLIENT_SECRET"),
		os.Getenv("MB_EMVI_ORGA"),
		nil)
	blog.nextUpdate = time.Now().Add(blogCacheTime)
	blog.loadArticles()
}

func (blog *Blog) loadArticles() {
	logbuch.Info("Refreshing blog articles...")
	articles, offset, count := make(map[string]emvi.Article), 0, 1
	var err error

	for count > 0 {
		var results []emvi.Article
		results, _, err = blog.client.FindArticles("", &emvi.ArticleFilter{
			BaseSearch:    emvi.BaseSearch{Offset: offset},
			Tags:          "blog",
			SortPublished: emvi.SortDescending,
		})

		if err != nil {
			logbuch.Error("Error loading blog articles", logbuch.Fields{"err": err})
			break
		}

		offset += len(results)
		count = len(results)

		for _, article := range results {
			articles[article.Id] = article
		}
	}

	if err == nil {
		for k, v := range articles {
			v.LatestArticleContent = blog.loadArticle(v)
			articles[k] = v
		}

		blog.setArticles(articles)
	}

	blog.nextUpdate = time.Now().Add(blogCacheTime)
	logbuch.Info("Done", logbuch.Fields{"count": len(articles)})
}

func (blog *Blog) loadArticle(article emvi.Article) *emvi.ArticleContent {
	_, content, _, err := blog.client.GetArticle(article.Id, article.LatestArticleContent.LanguageId, 0)

	if err != nil {
		logbuch.Error("Error loading article", logbuch.Fields{"err": err, "id": article.Id})
		return nil
	}

	logbuch.Debug("Article loaded", logbuch.Fields{"id": article.Id})
	return content
}

func (blog *Blog) setArticles(articles map[string]emvi.Article) {
	blog.articles = articles
	blog.articlesYear = make(map[int][]emvi.Article)

	for _, article := range articles {
		if blog.articlesYear[article.Published.Year()] == nil {
			blog.articlesYear[article.Published.Year()] = make([]emvi.Article, 0)
		}

		blog.articlesYear[article.Published.Year()] = append(blog.articlesYear[article.Published.Year()], article)
	}
}

func (blog *Blog) getArticle(id string) emvi.Article {
	blog.refreshIfRequired()
	return blog.articles[id]
}

func (blog *Blog) getArticles() map[int][]emvi.Article {
	blog.refreshIfRequired()
	return blog.articlesYear
}

func (blog *Blog) getLatestArticles() []emvi.Article {
	blog.refreshIfRequired()
	articles := make([]emvi.Article, 0, 3)
	i := 0

	for _, v := range blog.articles {
		articles = append(articles, v)
		i++

		if i > maxLatestArticles {
			break
		}
	}

	return articles
}

func (blog *Blog) refreshIfRequired() {
	if blog.nextUpdate.Before(time.Now()) {
		blog.loadArticles()
	}
}

func GetLatestArticles() []emvi.Article {
	return blog.getLatestArticles()
}

func ServeBlogPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Articles map[int][]emvi.Article
		}{
			blog.getArticles(),
		}

		if err := tpl.Get().ExecuteTemplate(w, "blog.html", data); err != nil {
			logbuch.Error("Error executing blog template", logbuch.Fields{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func ServeBlogArticle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		slug := strings.Split(vars["slug"], "-")

		if len(slug) == 0 {
			http.Redirect(w, r, "/notfound", http.StatusFound)
			return
		}

		article := blog.getArticle(slug[len(slug)-1])

		if len(article.Id) == 0 {
			http.Redirect(w, r, "/notfound", http.StatusFound)
			return
		}

		data := struct {
			Title     string
			Content   template.HTML
			Published string
		}{
			article.LatestArticleContent.Title,
			template.HTML(article.LatestArticleContent.Content),
			article.Published.Format("2. January 2006"),
		}

		if err := tpl.Get().ExecuteTemplate(w, "article.html", data); err != nil {
			logbuch.Error("Error executing blog article template", logbuch.Fields{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
