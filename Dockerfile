FROM golang:1.14-alpine AS build

WORKDIR /go/src/github.com/shu8/linkener

# build-base needed for gcc to build go-sqlite3
RUN apk add git build-base sqlite

COPY . .
RUN go mod download

ENV CGO_ENABLED=0
RUN go build -v ./...
RUN go install ./...

RUN ./setup.sh
RUN mkdir -p /var/lib/linkener
RUN mv auth.db /var/lib/linkener scratch
COPY --from=build /go/bin/linkener /
COPY --from=build /var/lib/linkener/ /var/lib/
EXPOSE 3000
ENTRYPOINT [ "/linkener" ]
