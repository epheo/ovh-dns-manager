FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o ovh-dns-manager \
    ./main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -s /bin/sh appuser

WORKDIR /app

COPY --from=builder /app/ovh-dns-manager .

RUN chown appuser:appuser ovh-dns-manager

USER appuser

ENTRYPOINT ["./ovh-dns-manager"]