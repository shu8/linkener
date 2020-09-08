FROM golang:1.14-alpine AS build

WORKDIR /go/src/github.com/shu8/url-shortener

# build-base needed for gcc to build go-sqlite3
RUN apk add git build-base

COPY . .
RUN go mod download

ENV CGO_ENABLED=0
RUN go build -v ./...
RUN go install ./...


FROM scratch
COPY --from=build /go/bin/url-shortener /
EXPOSE 3000
ENTRYPOINT [ "/url-shortener" ]
