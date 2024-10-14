# Goreleaser Dockerfile
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

COPY pasolo pasolo

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./pasolo"]