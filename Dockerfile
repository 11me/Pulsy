FROM golang:1.20-alpine AS builder
WORKDIR /opt/build
COPY . .
RUN go mod download && \
    go build -o pulsy .

FROM alpine:3.16.0
WORKDIR /opt/app
COPY --from=builder /opt/build/pulsy .
CMD ["./pulsy"]

