FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:3.19

RUN apk --no-cache add ca-certificates && \
    adduser -D -g '' appuser
USER appuser

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
