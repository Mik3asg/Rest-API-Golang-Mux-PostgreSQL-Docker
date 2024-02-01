# Use official Golang image
FROM golang:1.21.6-alpine3.19

# Set working directory
WORKDIR /app

# Copy source code
COPY . . 

# Download and install dependencies
RUN go get -d -v ./...

# Build GO app
RUN go build -o api .

# Expose port
EXPOSE 8000

# Run executable
CMD ["./api"]