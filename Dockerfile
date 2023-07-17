FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o tcp-state-exporter .

FROM alpine:latest  
WORKDIR /root/
COPY --from=builder /app/tcp-state-exporter .
CMD ["./tcp-state-exporter"]  
