# **Dependencies**
FROM golang:1.24.6 AS deps

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

# **Build App**
FROM golang:1.24.6 AS builder

WORKDIR /usr/src/app

COPY --from=deps /go/pkg /go/pkg
COPY . .

RUN go build -v -o /usr/local/bin/app cmd/main.go

# **Run Compiled App**
FROM golang:1.24.6 AS base

WORKDIR /usr/src/app

COPY --from=builder /usr/local/bin/app /usr/local/bin/app

EXPOSE 8080

CMD ["app"]
