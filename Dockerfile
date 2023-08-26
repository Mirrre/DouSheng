# Start from the latest golang base image
FROM golang:1.20

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .
COPY wait-for-it.sh /app/wait-for-it.sh

# Build the Go app
RUN go build -o main .

# Command to run the executable
CMD ["./main"]
