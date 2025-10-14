# services/web_crawler/Dockerfile
FROM golang:1.25 AS builder

WORKDIR /app/services/web_crawler

COPY libs/utils /app/libs/utils
COPY services/web_crawler/go.mod services/web_crawler/go.sum ./

RUN go mod tidy
RUN go mod download && go mod verify

COPY services/web_crawler ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /crawler main.go

FROM alpine:3.20

# Install CA certificates for HTTPS
RUN apk add --no-cache ca-certificates

COPY --from=builder /crawler /crawler

WORKDIR /app

ENTRYPOINT ["/crawler"]
