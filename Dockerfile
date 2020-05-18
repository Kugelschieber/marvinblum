FROM golang AS build
ADD . /go/src/github.com/special-tactical-service/wiki
WORKDIR /go/src/github.com/special-tactical-service/wiki
RUN apt-get update && \
	apt-get upgrade -y && \
	apt-get install -y curl && \
	curl -sL https://deb.nodesource.com/setup_13.x -o nodesource_setup.sh && bash nodesource_setup.sh && \
	apt-get install -y nodejs

# build backend
ENV GOPATH=/go
ENV CGO_ENABLED=0
RUN go build -ldflags "-s -w" main.go

# build frontend
RUN cd /go/src/github.com/special-tactical-service/wiki/public && npm i && npm rebuild node-sass && npm run build

FROM alpine
RUN apk update && \
    apk upgrade && \
    apk add --no-cache && \
    apk add ca-certificates && \
    rm -rf /var/cache/apk/*
COPY --from=build /go/src/github.com/special-tactical-service/wiki /app
WORKDIR /app

# default config
ENV STS_WIKI_LOGLEVEL=info
ENV STS_WIKI_ALLOWED_ORIGINS=*
ENV STS_WIKI_HOST=0.0.0.0:80

EXPOSE 80
EXPOSE 443
CMD ["/app/main"]
