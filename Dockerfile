# Use the Go 1.20 image as the builder stage
FROM golang:1.20-alpine as builder

WORKDIR /app

COPY . .

RUN apk add curl
# Build the Go application
RUN go build -o main cmd/main.go
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz

# Use the Alpine image for the final runtime
FROM alpine:3.16

WORKDIR /app

# Copy the compiled binary and other necessary files from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate
COPY views ./views
COPY templates ./templates
COPY static ./static
COPY migrations ./migrations
COPY package-lock.json ./package-lock.json
COPY package.json ./package.json
COPY tailwind.config.js ./tailwind.config.js

# Run the application
CMD ["./main"]
