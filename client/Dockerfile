# Start with a base image containing Go runtime
FROM golang:1.20 as builder
RUN apt-get update 
RUN apt-get install pkg-config libopus-dev libopusfile-dev -y
# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Build the application
RUN go build -o client cmd/faster-whisper/main.go

# Start a new stage from scratch
FROM golang:1.20

RUN apt-get update 
RUN apt-get install pkg-config libopus-dev libopusfile-dev -y
# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/client ./client

# Set the environment variable
ENV URL="localhost:8088"
ENV ROOM="test"
ENV TRASCRIPTION_SERVICE="localhost:8000"

# Run the binary program produced by `go install`
CMD ["./client"]
