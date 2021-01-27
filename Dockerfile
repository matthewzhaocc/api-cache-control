FROM golang:latest AS builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .

FROM alpine:latest
COPY --from=0 /build/api-cache-control .
CMD ./api-cache-control