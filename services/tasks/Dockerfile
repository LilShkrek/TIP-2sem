FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/tasks ./cmd/tasks

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /out/tasks /app/tasks

EXPOSE 8080

ENTRYPOINT ["/app/tasks"]
