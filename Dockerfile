# **Dependencies**
FROM golang:1.26.1 AS deps

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4

# **Image for Dev things**
FROM deps AS dev

WORKDIR /usr/src/app

COPY --from=deps /go/pkg /go/pkg
COPY . .
COPY migrations /migrations
COPY ./config/stripe_prices.yml /config/stripe_prices.yml

# **Build App**
FROM deps AS builder

WORKDIR /usr/src/app

COPY . .

RUN swag init --generalInfo cmd/main.go --output ./swagger

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /usr/local/bin/app cmd/main.go

# **Run test for Compiled App**
FROM builder AS test

WORKDIR /usr/src/app

COPY migrations /migrations
COPY ./config/stripe_prices.yml /config/stripe_prices.yml
COPY ./scripts/run_tests.sh ./run_tests.sh

RUN chmod +x ./run_tests.sh

CMD ["/usr/local/bin/app"]

# **Run Compiled App**
FROM gcr.io/distroless/base-debian12 AS prod

WORKDIR /usr/src/app

COPY certs /certs
COPY migrations /migrations
COPY ./config/stripe_prices.yml /config/stripe_prices.yml
COPY templates ./templates
COPY internal/world-service/nomad/ftr-server-job.nomad /nomad/templates/ftr-server-job.nomad
COPY --from=builder /usr/local/bin/app /usr/local/bin/app

USER nonroot:nonroot

EXPOSE 8000

ENTRYPOINT ["/usr/local/bin/app"]
