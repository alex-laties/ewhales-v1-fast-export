# Stage 1: Build the Go application
FROM golang:alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
# If go.mod has a version newer than the alpine image, it might complain. 
# We can safely adjust the version or just download.
RUN go mod download

# Copy source code
COPY *.go ./

# Build the exporter binary statically
RUN CGO_ENABLED=0 go build -o exporter .

# Stage 2: Final lightweight image
FROM python:3.11-alpine

WORKDIR /app

# Copy the compiled Go binary
COPY --from=builder /app/exporter /usr/local/bin/exporter

# Copy the Python scripts
COPY convert_inserts.py /usr/local/bin/convert_inserts.py
COPY anonymize.py /usr/local/bin/anonymize.py

# Ensure scripts are executable
RUN chmod +x /usr/local/bin/exporter /usr/local/bin/convert_inserts.py /usr/local/bin/anonymize.py

# Default entrypoint to interactive shell
CMD ["/bin/sh"]
