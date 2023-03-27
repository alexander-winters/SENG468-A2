# Use the official Golang image as the base image
FROM golang:latest

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files to the container
COPY ../go.mod ../go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code to the container
COPY . .

# Build the application
RUN go build -o main ./server

# Expose the port on which the app will run
EXPOSE 3000

# Run the binary
CMD ["./main"]