# Start with a base image containing Go runtime
FROM golang:latest as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Start a new stage from scratch
FROM golang:1.20  


# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/main .
COPY --from=builder /app/config.toml .
COPY --from=builder /app/web ./web

EXPOSE 8088

# Run the binary program produced by `go install`
CMD ["./main"]