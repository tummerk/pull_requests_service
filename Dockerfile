FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s" -o /app/server ./cmd/main.go


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/.env .
COPY --from=builder /app/db ./db

EXPOSE 8080

CMD ["./server"]