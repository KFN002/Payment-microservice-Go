FROM golang:1.23.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /app/payment-service ./main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y libc6 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/payment-service /payment-service

EXPOSE 50051

CMD ["/payment-service"]
