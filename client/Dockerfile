# Start with a base image containing Go runtime
FROM golang:1.20 as builder
RUN apt-get update 
RUN apt-get install pkg-config libopus-dev libopusfile-dev -y
# Set the Current Working Directory inside the container
WORKDIR /app

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY ./log/. ./log/.
COPY ./stt/. ./stt/.
COPY ./tts/backends/. ./tts/backends/.
COPY ./tts/engine/. ./tts/engine/.
COPY ./client/. ./client/.
COPY ./rtc/. ./rtc/.
COPY ./util/. ./util/.
COPY ./go.work .


WORKDIR /app/client

# Build the application
RUN go build -o client cmd/http/main.go

# Start a new stage from scratch
FROM golang:1.20

RUN apt-get update 
RUN apt-get install pkg-config libopus-dev libopusfile-dev -y
# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/client/client ./client

# Set the environment variable
ENV URL="localhost:8088"
ENV ROOM="test"
ENV TRASCRIPTION_SERVICE="localhost:8000"

# Run the binary program produced by `go install`
CMD ["./client"]
