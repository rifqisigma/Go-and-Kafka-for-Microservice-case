# Gunakan base image Golang
FROM golang:1.24-alpine

# Set working directory di dalam container
WORKDIR /app

# Copy semua file ke working dir container
COPY . .

# Unduh dependencies
ENV GOPROXY=https://proxy.golang.org,direct
RUN go mod download


# Build binary app
RUN go build -o main ./cmd/main.go

# Expose port jika ada HTTP server
EXPOSE 3004

# Jalankan aplikasi
CMD ["./main"]