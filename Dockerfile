FROM golang:1.14.2 as builder
COPY . /build
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -o ddns

#second stage
FROM alpine:latest
WORKDIR /root/
RUN apk --update add ca-certificates
COPY --from=builder /build/ddns .
CMD ["./ddns"]