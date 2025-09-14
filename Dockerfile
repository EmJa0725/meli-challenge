# Stage 1: Build
FROM golang:1.25-bookworm as builder
WORKDIR /app
# Copy go mod and sum files for caching
COPY go.mod go.sum ./
RUN go mod download
# Copy the source code
COPY . .
# Build the application
RUN go build -o meli-challenge .

# Stage 2: Run
FROM debian:bookworm-slim
# Cretate a non-root user and group
RUN groupadd -r appgroup && useradd -r -g appgroup appuser
WORKDIR /app
# Copy the built binary from the builder stage
COPY --from=builder /app/meli-challenge .
# Change permissions
RUN chown -R appuser:appgroup /app
# Change to a non-root user
USER appuser
# Expose the application port
EXPOSE 8080
# Command to run the application
CMD ["./meli-challenge"]