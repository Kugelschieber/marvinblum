FROM golang AS build
COPY . /go/src/github.com/Kugelschieber/marvinblum.de
WORKDIR /go/src/github.com/Kugelschieber/marvinblum.de
RUN apt-get update && apt-get upgrade -y

ENV GOPATH=/go
ENV CGO_ENABLED=0
RUN go build -ldflags "-s -w" main.go

FROM alpine
RUN apk update && \
    apk upgrade && \
    apk add --no-cache && \
    apk add ca-certificates && \
    rm -rf /var/cache/apk/*
COPY --from=build /go/src/github.com/Kugelschieber/marvinblum.de /app
WORKDIR /app

# default config
ENV MB_LOGLEVEL=info
ENV MB_ALLOWED_ORIGINS=*
ENV MB_HOST=0.0.0.0:8888
ENV MB_DB_PORT=5432
ENV MB_DB_SSLMODE=disable

EXPOSE 8888
CMD ["/app/main"]
