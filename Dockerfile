FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o coral-watchdog .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/coral-watchdog .
COPY --from=builder /app/frontend ./frontend
COPY --from=builder /app/sources ./sources
COPY --from=builder /app/queries ./queries

# NOTE: coral.exe must be added separately — it's not in this image
# Download from https://github.com/withcoral/coral/releases
# and mount it: docker run -v ./coral:/app/coral ...

EXPOSE 8080
CMD ["./coral-watchdog"]