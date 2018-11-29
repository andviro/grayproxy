FROM golang:1.11-alpine AS builder
RUN apk add --update alpine-sdk
WORKDIR /tmp/app
COPY . .
RUN go build -mod=vendor -o grayproxy

FROM alpine:latest
RUN apk add --update ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /tmp/app/grayproxy /grayproxy
EXPOSE 12201/udp

CMD ["/grayproxy"]
