FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o migration ./cmd/migration

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=UTC

WORKDIR /root/

COPY --from=builder /app/migration .
COPY --from=builder /app/migrations ./migrations

CMD ["./migration"]

