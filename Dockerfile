FROM golang:1.20.11 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download && apt update && apt install -y sqlite3 

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app with CGO enabled
RUN go get github.com/mattn/go-sqlite3 && CGO_ENABLED=1 GOOS=linux go build  -installsuffix cgo -o gpt-go .

CMD ["./gpt-go"]
