FROM golang:1.20-alpine as builder
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates
WORKDIR /app
COPY . .
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -mod=vendor -ldflags="-w -s" -o main

FROM scratch as production
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/main /grayproxy
EXPOSE 12201/udp
ENTRYPOINT ["/grayproxy"]
CMD ["-in", "udp://:12201", "-v"]
