FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache wget && adduser -D -u 10001 appuser

WORKDIR /app

COPY --from=builder /app/server /app/server
COPY --from=builder /app/api/openapi.yaml /app/api/openapi.yaml

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/server"]
