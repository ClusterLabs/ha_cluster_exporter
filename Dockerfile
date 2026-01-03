# Build stage
FROM golang:1.24.11-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ha_cluster_exporter ./cmd/ha_cluster_exporter

# Final stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/ha_cluster_exporter .
# Copy sample config
COPY ha_cluster_exporter.yaml /etc/ha_cluster_exporter.yaml

# Expose port
EXPOSE 9664

# Command to run
ENTRYPOINT ["./ha_cluster_exporter"]
CMD ["--web.listen-address=:9664"]
