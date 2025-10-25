FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o mockserver cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/mockserver .
COPY --from=builder /app/etc ./etc

EXPOSE 8888

CMD ["./mockserver", "-f", "etc/mockserver.yaml"]
