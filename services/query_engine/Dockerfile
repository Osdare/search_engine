FROM golang:1.25 AS builder

WORKDIR /app/services/query_engine

COPY libs/utils /app/libs/utils
COPY services/query_engine/go.mod services/query_engine/go.sum ./

RUN go mod tidy
RUN go mod download && go mod verify

COPY services/query_engine ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /query_engine main.go

FROM alpine:3.20

# Install CA certificates for HTTPS
RUN apk add --no-cache ca-certificates

COPY --from=builder /query_engine /query_engine

WORKDIR /app

ENTRYPOINT ["/query_engine"]

