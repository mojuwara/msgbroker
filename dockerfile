# syntax=docker/dockerfile:1

# Go version 1.17.8 running on Alpine Linux
FROM golang:1.17.8-alpine

# Work in a subdir called app
WORKDIR /app

# Copy source files into container file system
COPY go.mod ./
COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy source code into the container
COPY *.go ./

# Compile our application and store binary in the root
RUN go build -o /msgbroker

# Recommended port
EXPOSE 3000

# Finally run the executable
CMD ["/msgbroker"]
