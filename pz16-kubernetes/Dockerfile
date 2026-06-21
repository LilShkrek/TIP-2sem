FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/tasks ./cmd/tasks

FROM scratch

COPY --from=builder /out/tasks /tasks

USER 65532:65532
EXPOSE 8082

ENTRYPOINT ["/tasks"]
