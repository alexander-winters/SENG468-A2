# Use the official Golang image as the base image
FROM golang:latest as builder

# Set the working directory
WORKDIR /app

# Copy the go.mod and go.sum files from the root directory to the current working directory
COPY ../go.mod ../go.sum ./

# Download the dependencies
RUN go mod download

# Copy the rest of the source code
COPY ./ ./

# Build the Go app
RUN go build -o main ./server

# Start a new stage with the base image
FROM golang:latest

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main /app/main

# Expose the port on which the app will run
EXPOSE 3000

# Run the app
CMD ["/app/main"]