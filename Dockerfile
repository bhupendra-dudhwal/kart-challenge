FROM golang:1.24.6-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARC=amd64 \
    go build -o migratiom ./cmd/migration && \
    go build -o kart ./cmd/http

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/kart .
COPY --from=builder /app/migratiom .

CMD ["sh" ,"-c", "./migratiom && ./kart"]