# Start with a base image containing Go runtime
FROM golang:latest as builder


WORKDIR /app/rtc
# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY ./rtc/. .


# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Start a new stage from scratch
FROM golang:1.20  
WORKDIR /app
# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/rtc/main ./rtc/main
COPY --from=builder /app/rtc/config.toml ./rtc/config.toml
COPY ../web/. ./web/.
WORKDIR /app/rtc
EXPOSE 8088
# Run the binary program produced by `go install`
CMD ["./main"]
