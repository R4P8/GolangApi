FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o go-example-api ./cmd

FROM ubuntu:22.04

WORKDIR /app

COPY --from=builder /app/go-example-api .

EXPOSE 8080


CMD ["./go-example-api"]