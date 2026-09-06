FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /loadtest-server ./tests/load/testapp

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /loadtest-server /loadtest-server
EXPOSE 8080
CMD ["/loadtest-server"]
