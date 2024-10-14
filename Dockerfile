# Stage 1: Build the Remix frontend
FROM node:20-alpine AS frontend-builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy package.json and package-lock.json
COPY frontend/package.json frontend/package-lock.json ./

# Install dependencies
RUN npm install

# Copy the source from the current directory to the Working Directory inside the container
COPY frontend/ .

# Build the Remix app
RUN npm run build

# Stage 2: Build the Go server
FROM golang:1.23-alpine AS server-builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY ./ .

# Copy the Remix build output from the frontend-builder stage
COPY --from=frontend-builder /app/build ./frontend/build

# Build the Go app
# Ensure the Remix build output exists
RUN if [ ! -f frontend/build/client/index.html ]; then echo "Remix build output not found!"; exit 1; fi

# Build the Go app
RUN go build -o server .

# Stage 3: Create the final image
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go server binary from the server-builder stage
COPY --from=server-builder /app/server .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./server"]