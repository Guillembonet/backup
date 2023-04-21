# Build the service to a binary
FROM golang:1.20.3-alpine AS builder

# Install packages
RUN apk add --no-cache bash gcc musl-dev linux-headers git

# Compile application
WORKDIR /go/src/github.com/guillembonet/backup
ADD . .
RUN go build -o build/main main.go

# Copy and run the made binary
FROM alpine:3.17

COPY --from=builder /go/src/github.com/guillembonet/backup/build/main /usr/bin/backup

ENTRYPOINT ["/usr/bin/backup"]