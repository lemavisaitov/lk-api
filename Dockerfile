FROM golang:1.23-alpine AS builder

WORKDIR /

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/api-gateway ./cmd/
RUN ls -l /app

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/api-gateway .

ENTRYPOINT ["./api-gateway"]