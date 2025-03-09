FROM golang:1.23-alpine AS builder

WORKDIR /

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/api-gateway ./cmd/
RUN ls -l /app

FROM alpine:latest

WORKDIR /app
COPY .env .
COPY ./migrations .
COPY --from=builder /app/api-gateway .

ENTRYPOINT ["./api-gateway"]