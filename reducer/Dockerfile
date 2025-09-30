FROM golang:1.25-alpine

WORKDIR /app

# Copy everything from reducer directory
COPY . .

# Download dependencies
RUN go mod download

# Build the application
RUN go build -o reducer .

EXPOSE 8080

CMD ["./reducer"]