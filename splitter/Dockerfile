FROM golang:1.25-alpine

WORKDIR /app

# Copy everything from splitter directory
COPY . .

# Download dependencies
RUN go mod download

# Build the application
RUN go build -o splitter .

EXPOSE 8080

CMD ["./splitter"]