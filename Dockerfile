# Start from the latest golang base image
FROM golang:1.20

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the wait-for-it script to the container
COPY wait-for-it.sh /app/wait-for-it.sh

# Ensure the script is executable
RUN chmod +x /app/wait-for-it.sh

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Command to run the Go application directly (instead of the compiled binary)
CMD ["go", "run", "."]
