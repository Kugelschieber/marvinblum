package blog

import (
	"fmt"
	"github.com/Kugelschieber/marvinblum.de/tpl"
	emvi "github.com/emvi/api-go"
	"github.com/emvi/logbuch"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	blogCacheTime     = time.Hour
	blogFileCache     = "static/blog"
	maxLatestArticles = 3
)

var (
	blog               Blog
	linkRegex          = regexp.MustCompile(`(?iU)href="/read/([^"]+)"`)
	attachmentRegex    = regexp.MustCompile(`(?iU)(href|src)="([^"]+)/api/v1/content/([^"]+)"`)
	attachmentURLRegex = regexp.MustCompile(`(?iU)(href|src)="([^"]+/api/v1/content/)([^"]+)"`)
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

	if err := os.MkdirAll(blogFileCache, 0755); err != nil {
		logbuch.Error("Error creating blog file cache directory", logbuch.Fields{"err": err})
	}

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

	blog.downloadAttachments(article.Id, content.Content)
	content.Content = linkRegex.ReplaceAllString(content.Content, `href="/blog/$1"`)
	content.Content = attachmentRegex.ReplaceAllString(content.Content, fmt.Sprintf(`$1="/static/blog/%s/$3"`, article.Id))
	logbuch.Debug("Article loaded", logbuch.Fields{"id": article.Id})
	return content
}

func (blog *Blog) downloadAttachments(id, content string) {
	if _, err := os.Stat(filepath.Join(blogFileCache, id)); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Join(blogFileCache, id), 0755); err != nil {
			logbuch.Error("Error creating article file cache directory", logbuch.Fields{"err": err, "id": id})
			return
		}
	}

	results := attachmentURLRegex.FindAllStringSubmatch(content, -1)

	for _, attachment := range results {
		if len(attachment) == 4 {
			resp, err := http.Get(attachment[2] + attachment[3])

			if err != nil {
				logbuch.Error("Error downloading blog attachment", logbuch.Fields{"err": err, "id": id})
				continue
			}

			defer resp.Body.Close()
			data, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				logbuch.Error("Error reading blog attachment body", logbuch.Fields{"err": err, "id": id})
				continue
			}

			if err := ioutil.WriteFile(filepath.Join(blogFileCache, id, attachment[3]), data, 0755); err != nil {
				logbuch.Error("Error saving blog attachment on disk", logbuch.Fields{"err": err, "id": id})
			}
		}
	}
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
