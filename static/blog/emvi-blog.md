**Published on 14. June 2020**

Welcome to my blog! My name is Marvin, I'm a software engineer and entrepreneur. I write about programming, servers, my work and everything I'm interested in. In my first blog post, I would like to show you how I build my website and the tools I used. You already guessed that from the title I suppose.

You can find the full source code for my website on [GitHub](https://github.com/Kugelschieber/marvinblum). It's MIT licensed, so you can build your own on top of it or just reuse parts of the code.

Goals
-----

So first of all, here are the goals I set when I started:

*   the page must be self hosted, I do like to have full control
*   it must be fast and have a small footprint
*   easy to deploy on cheap hardware
*   I don't want to put too much thought and time into styling
*   enable me to write articles without having to change the page itself, but don't require me to install (and update!) a full fledged CMS at the same time
    

The last point is probably the most important one to me. My page won't change very frequently and I don't want to maintain a CMS. I also don't want to write a template for any CMS out there, as that quickly gets out of hand and is not worth the effort. Static HTML won't do it neither, as the blog articles need to be updated as soon as I release a new one or change an existing one.

Lets go through the bullet points and the choices I made. The most interesting part is probably how I build the blog.

Server and Deployment
---------------------

For hosting, I chose [Hetzner](https://www.hetzner.com/) as a cloud provider. The Hetzner cloud offers virtual machines, block storage and networking (subnets, floating IPs, ...). There is an API too, which can be used to automate things.

My website is hosted on the smallest VM (CX11-CEPH) for 2,96 €/month, which is insanely cheap. It provides a single vCPU, 2GB RAM and 20GB storage. Which is sufficient for my simple page. I chose a CEPH machine, as this will store all data on block storage rather than on the machine itself, which decouples it from the hardware. In case of a hardware fault, Hetzner will boot up my server on a different machine and I won't have to do anything. I'm not sure if it assigns a different IP to the server in that case. For the OS I chose Ubuntu as I use that on my computer and I'm familiar with it.

The software running my page is a custom server I build using [Go](https://golang.org/) (golang), as it is an excellent programming language and offers high performance. I will go into more detail about the code in a second.

I use Docker and Compose to deploy my page. Both are well established tools to package and deploy software. These are the only tools I installed on the VM, so I just need to update the systems packages through _apt_ from time to time. Within the `docker-compose.yml` I added [Traefik](https://containo.us/traefik/) as a reverse proxy to schedule a SSL certificate from Letsencrypt.

Deploying my page is now as simple as building and pushing the Docker image, pulling it on the server and restarting the container. Of course you could automate that whole process so that the page updates itself, but again: I won't change the content frequently. So that's good enough.

Structure and Static Content
----------------------------

Lets taking a look at the directory structure:

*   blog - code to load and cache blog articles
*   static - static files (my picture, stylesheets, ...) and used to cache blog article attachments (more on that later)
*   template - contains the HTML files to build the page
*   tpl - code to load and build the page from the template files
    

The root directory contains the `main.go` to wire everything up and set up the router, as well as the `Dockerfile`, `docker-compose.yml` and the Go dependencies (`go.mod`). Everything within the `static` directory is served as static content on the `/static/...` route. Each page has it's own handler function which assembles the HTML using the template files.

Another point worth mentioning is gzip compression. I added the `gziphandler.GzipHandler` on the static route to compress files. The middleware is build by the New York Times and easy to integrate. You can check it out [here](https://github.com/nytimes/gziphandler).

Styling
-------

As I do like to keep things simple, I chose a micro CSS framework so that I don't have to bother with styling too much. Namely [concrete](https://concrete.style/), which I adjusted a bit, to narrow the layout and add a header with my picture. Apart from that I'm quite pleased with the look of it. As a bonus, it also switches to dark mode automatically if you set that in your (OS) preferences.

Templating
----------

To prevent writing the same HTML over and over again I made use of Go's template system. It's simple but powerful enough for most websites and you can extend it using function maps. Here is an example for the blog article page (the one you're looking at right now):

```
{{"{{"}}template "head.html"{{"}}"}}
{{"{{"}}template "menu.html"{{"}}"}}

<section>
    <h1>{{"{{"}}.Title{{"}}"}}</h1>
    <small>Published on {{"{{"}}format .Published "2. January 2006"{{"}}"}}</small>
    {{"{{"}}.Content{{"}}"}}
</section>

{{"{{"}}template "end.html"{{"}}"}}
```

`head`, `menu` and `end` are reused on all pages.

I've added two functions to format dates and build the blog article slug from the title:

```
var funcMap = template.FuncMap{
	"slug": slug.Make,
	"format": func(t time.Time, layout string) string {
		return t.Format(layout)
	},
}
```

Blog
----

[Emvi](https://emvi.com/) offers an API which allows anyone to use it as a headless CMS. The main advantage of it is, that I can use its editor to write my blog articles, upload images/files and don't need to worry about hosting my own CMS. Apart from that I'm using Emvi for note taking and documentation anyways, so I can stay on the same platform.

To read articles, I make use of the [Go client library](https://github.com/emvi/api-go). It isn't complete yet, as Emvi is still in beta, but provides everything required to build a blog. On top of it I build my own type to cache articles and files and sort them into maps, which are rendered on my page later. You could just use the client to do all of that without caching, but to reduce latency and serve articles in case Emvi goes down for some reason, I thought that would be a good idea.

```
type Blog struct {
	client       *emvi.Client
	articles     map[string]emvi.Article // id -> article
	articlesYear map[int][]emvi.Article  // year -> articles
	nextUpdate   time.Time
}
```

The `client` is initialized with the client ID and secret I generated within Emvi, as well as the name of my organization. These are configured using environment variables, so that I can put them into the `docker-compose.yml`. `nextUpdate` is used to refresh the cache after some time. Articles and files will only be updated in case they have changed since the last time they have been accessed. The article content itself is cached in memory, files are stored on disk.

Articles are put into two different maps. The first one is used to access any article by ID. The ID is read from the slug within the URL to render an article. The second map groups all articles by year, which is used to display them on the blogs overview page.

Note that you need to set an article to "external" within Emvi to allow it to be read through the API. To prevent reading articles which do not belong to my blog, I filtered the results by the tag "blog" and sort them in descending order:

```
results, _, err = blog.client.FindArticles("", &emvi.ArticleFilter{
	BaseSearch:    emvi.BaseSearch{Offset: offset},
	Tags:          "blog",
	SortPublished: emvi.SortDescending,
})
```

The offset is provided to read articles in a loop, as you can only read a fixed amount of results in one call. Afterwards, the content and files are read and cached for all results. I also added some regex to replace the paths within the content of each article to read images and files from my page instead of accessing Emvi.

And that's pretty much it. If you now visit my website, it will extract the ID from the URL, look up the cache, update it if required and return the result to you.

Conclusion
----------

Personal blogging is something I love about the internet and I now started my own blog. In terms of cost, running this page costs me 2,96 €/month for the server and 5$/month for Emvi (also I'm not paying for it as I'm the co-founder) plus something for the domain, which is insignificant. The solution I chose is fun and easy to implement, but certainly not suitable for non-programmers. I hope I can provide a plug and play solution in the future. It will most likely also use Emvi, as we are turning it into a platform for all sorts of different applications.

In case you would like to send me feedback or have a question, you can contact me by [mail](mailto:marvin@marvinblum.de) or on [Twitter](https://twitter.com/m5blum).
