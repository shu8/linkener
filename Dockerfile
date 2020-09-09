FROM golang:1.14-alpine AS build

WORKDIR /go/src/github.com/shu8/url-shortener

# build-base needed for gcc to build go-sqlite3
RUN apk add git build-base sqlite

COPY . .
RUN go mod download

ENV CGO_ENABLED=0
RUN go build -v ./...
RUN go install ./...

RUN ./setup.sh
RUN mkdir -p /var/lib/url-shortener
RUN mv auth.db /var/lib/url-shortener

FROM scratch
COPY --from=build /go/bin/url-shortener /
COPY --from=build /var/lib/url-shortener/ /var/lib/
EXPOSE 3000
ENTRYPOINT [ "/url-shortener" ]
