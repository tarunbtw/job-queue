#  Build stage 
FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o server ./cmd/server
RUN CGO_ENABLED=0 go build -o worker ./cmd/worker

# Runtime
FROM alpine:latest
WORKDIR /app

RUN mkdir -p /app/data

# Binaries
COPY --from=builder /app/server  ./server
COPY --from=builder /app/worker  ./worker

# Dashboard static files
COPY dashboard/ ./dashboard/

# Entrypoint
COPY entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

EXPOSE 8080

# Both server and worker share /app/data/jobs.db on the local filesystem.
# Multiple instances would each maintain a separate database file,
# splitting job history across instances.
ENTRYPOINT ["./entrypoint.sh"]
