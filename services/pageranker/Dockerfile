# services/pageranker/Dockerfile
FROM golang:1.25 AS builder

WORKDIR /app/services/pageranker

COPY libs/utils /app/libs/utils
COPY services/pageranker/go.mod services/pageranker/go.sum ./

RUN go mod tidy
RUN go mod download && go mod verify

COPY services/pageranker ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /pageranker main.go

FROM alpine:3.20

# Install CA certificates for HTTPS
RUN apk add --no-cache ca-certificates

COPY --from=builder /pageranker /pageranker

WORKDIR /app

ENTRYPOINT ["/pageranker"]
