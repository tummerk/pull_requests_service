FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git make
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download


COPY . .

RUN make generate
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -o /app/main .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/db/migrations ./db/migrations

COPY --from=builder /go/bin/migrate .

COPY --from=builder /app/main .

COPY docker/entrypoint.sh .
RUN chmod +x ./entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["./entrypoint.sh"]
CMD ["./main"]