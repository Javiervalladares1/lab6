
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk update && apk add --no-cache git ca-certificates


COPY go.mod go.sum ./
RUN go mod download


COPY . .
RUN go build -o laliga-tracker .


FROM alpine:latest

WORKDIR /app


COPY --from=builder /app/laliga-tracker .


EXPOSE 8080

CMD ["./laliga-tracker"]