# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from golang:1.12-alpine base image
FROM golang:1.23-alpine

# The latest alpine images don't have some tools like (`git` and `bash`).
# Adding git, bash and openssh to the image
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

# Add Maintainer Info
LABEL maintainer="Rajeev Singh <rajeevhub@gmail.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependancies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main ./cmd/app

# Copy and make the entrypoint script executable
COPY golang/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose port 8080 to the outside world
EXPOSE 8080

# Run the executable
ENTRYPOINT ["/entrypoint.sh"]