FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /tps ./cmd/server

FROM alpine:latest

COPY --from=builder /tps /tps

EXPOSE 8080

CMD ["/tps"]