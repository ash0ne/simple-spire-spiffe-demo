FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY . .

RUN go mod init demo && \
    go mod tidy && \
    go build -o /client ./client.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /client .

ENTRYPOINT ["./client"]
