# Start with a base image containing Go runtime
FROM golang:1.20 as builder
RUN apt-get update 
RUN apt-get install pkg-config libopus-dev libopusfile-dev -y
# Set the Current Working Directory inside the container
WORKDIR /app

COPY ./. .

# Build the application
RUN go build -o client cmd/http/main.go

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
