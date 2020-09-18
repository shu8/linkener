FROM golang:1.14-alpine AS build

WORKDIR /go/src/github.com/shu8/linkener

# build-base needed for gcc to build go-sqlite3
RUN apk add git build-base sqlite

COPY . .
RUN go mod download

RUN go build -v ./... && go install ./...

RUN ./setup.sh
RUN mkdir -p /var/lib/linkener && mv auth.db /var/lib/linkener/
RUN mkdir -p .linkener && mv config.json .linkener/

FROM alpine

# Linkener binary executable
COPY --from=build /go/bin/linkener /
# Auth DB
COPY --from=build /var/lib/linkener/ /var/lib/linkener/
# Default config
COPY --from=build /go/src/github.com/shu8/linkener/.linkener/ /root/.linkener/

EXPOSE 3000
ENTRYPOINT [ "/linkener" ]
