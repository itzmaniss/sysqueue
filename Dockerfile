FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o sysqueue .

FROM alpine:latest
COPY --from=builder /app/sysqueue /sysqueue
ENTRYPOINT ["/sysqueue"]
