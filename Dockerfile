FROM golang:1.23-alpine AS builder

WORKDIR /

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/api-gateway ./cmd/
RUN ls -l /app

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/api-gateway .

ADD migrations/*.sql migrations/
ADD migration.sh .
ADD .env .
ADD https://github.com/pressly/goose/releases/download/v3.24.1/goose_linux_x86_64 /bin/goose
RUN chmod +x /bin/goose
RUN chmod +x migration.sh

ENTRYPOINT ["./api-gateway"]