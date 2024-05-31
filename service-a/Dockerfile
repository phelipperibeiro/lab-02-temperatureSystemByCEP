# Use an official Golang runtime as a parent image
FROM golang:1.22 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Download dependencies
RUN go mod download

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main cmd/main.go

# Use a minimal Docker image to run the Go app
FROM scratch

# Copy the compiled Go binary from the builder image
COPY --from=builder /app/main /app/main

# Expose the port the app runs on
EXPOSE 8080

# Run the Go binary
ENTRYPOINT ["/app/main"]


# docker network create lab2_network
# docker build -f Dockerfile -t service-a . 
# docker run -d -p 8080:8080 --name service-a --network lab2_network service-a

