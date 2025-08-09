FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s' \
    -o ovh-dns-manager \
    ./main.go

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/ovh-dns-manager /ovh-dns-manager

ENTRYPOINT ["/ovh-dns-manager"]