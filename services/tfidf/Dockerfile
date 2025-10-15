FROM golang:1.25 AS builder

WORKDIR /app/services/tfidf

COPY libs/utils /app/libs/utils
COPY services/tfidf/go.mod services/tfidf/go.sum ./

RUN go mod tidy
RUN go mod download && go mod verify

COPY services/tfidf ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /tfidf main.go

FROM alpine:3.20

# Install CA certificates for HTTPS
RUN apk add --no-cache ca-certificates

COPY --from=builder /tfidf /tfidf

WORKDIR /app

ENTRYPOINT ["/tfidf"]

