# **Dependencies**
FROM golang:1.24.6 AS deps

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

# **Build App**
FROM deps AS builder

WORKDIR /usr/src/app

COPY --from=deps /go/pkg /go/pkg
COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init --generalInfo cmd/main.go

RUN go build -v -o /usr/local/bin/app cmd/main.go

# **Run Compiled App**
FROM builder AS prod

WORKDIR /usr/src/app

COPY --from=builder /usr/local/bin/app /usr/local/bin/app

EXPOSE 8080

CMD ["app"]

# **Run test for Compiled App**
FROM builder AS test

WORKDIR /usr/src/app

COPY --from=builder /usr/local/bin/app /usr/src/app/app

RUN chmod +x run_tests.sh

CMD ["app"]

# **Image for Dev things**
FROM deps AS dev

WORKDIR /usr/src/app

COPY --from=deps /go/pkg /go/pkg
COPY . .
