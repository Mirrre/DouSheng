# Start from the latest golang base image
FROM golang:1.20

# Set the Current Working Directory inside the container
WORKDIR /app

# Change the APT source to ustc mirrors
RUN echo "deb https://mirrors.ustc.edu.cn/debian/ bookworm main contrib non-free non-free-firmware" > /etc/apt/sources.list && \
    echo "deb-src https://mirrors.ustc.edu.cn/debian/ bookworm main contrib non-free non-free-firmware" >> /etc/apt/sources.list && \
    echo "deb https://mirrors.ustc.edu.cn/debian/ bookworm-updates main contrib non-free non-free-firmware" >> /etc/apt/sources.list && \
    echo "deb-src https://mirrors.ustc.edu.cn/debian/ bookworm-updates main contrib non-free non-free-firmware" >> /etc/apt/sources.list && \
    echo "deb-src https://mirrors.ustc.edu.cn/debian/ bookworm-updates main contrib non-free non-free-firmware" >> /etc/apt/sources.list && \
    echo "deb https://mirrors.ustc.edu.cn/debian/ bookworm-backports main contrib non-free non-free-firmware" >> /etc/apt/sources.list && \
    echo "deb-src https://mirrors.ustc.edu.cn/debian/ bookworm-backports main contrib non-free non-free-firmware" >> /etc/apt/sources.list && \
    echo "deb https://mirrors.ustc.edu.cn/debian-security/ bookworm-security main contrib non-free non-free-firmware" >> /etc/apt/sources.list && \
    echo "deb-src https://mirrors.ustc.edu.cn/debian-security/ bookworm-security main contrib non-free non-free-firmware" >> /etc/apt/sources.list

# Update package list and install ffmpeg
RUN apt-get update && apt-get install -y ffmpeg && rm -rf /var/lib/apt/lists/*

# Copy the wait-for-it script to the container
COPY wait-for-it.sh /app/wait-for-it.sh

# Ensure the script is executable
RUN chmod +x /app/wait-for-it.sh

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Command to run the Go application directly (instead of the compiled binary)
CMD ["go", "run", "."]
