FROM golang:1.24-alpine

# Install bash (git tidak diperlukan)
RUN apk add --no-cache bash

# Set working directory
WORKDIR /app

# Copy dan download dependency terlebih dahulu (layer cache lebih efisien)
COPY go.mod go.sum ./
ENV GOPROXY=https://proxy.golang.org,direct
RUN go mod download

# Copy seluruh source code
COPY . .

# Build binary untuk migrate dan main
RUN go build -o migrate ./cmd/migrate/main.go
RUN go build -o main ./cmd/main.go


# run port http server
EXPOSE 3002

# Jalankan migrate dulu, lalu main
CMD ["sh", "-c", "./migrate && ./main"]
