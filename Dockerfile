FROM golang:1.25.4-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /gocrawl \
    ./cmd/gocrawl

FROM alpine:latest

RUN apk add --no-cache ca-certificates && \
    update-ca-certificates

COPY --from=builder /gocrawl /usr/local/bin/gocrawl

ENTRYPOINT ["gocrawl"]
