FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY . .

RUN go mod init demo && \
    go mod tidy && \
    go build -o /server ./server.go

FROM alpine:latest

RUN apk add --no-cache curl ca-certificates

WORKDIR /app

COPY --from=builder /server .

ENTRYPOINT ["./server"]
