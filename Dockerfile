FROM alpine:latest
ENV GOPATH=/go \
    PATH=$PATH:$GOPATH/bin
RUN mkdir -p /go/src/github.com/andviro/grayproxy
ADD . /go/src/github.com/andviro/grayproxy
RUN apk update && \
    apk add git go libc-dev && \
    go get -v github.com/andviro/grayproxy/...  && \
    apk del git go libc-dev && \
    rm -rf /go/src && \
    rm -rf /go/pkg && \
    rm -rf /var/cache/apk/*
EXPOSE 12201/udp
ENTRYPOINT ["/go/bin/grayproxy"]
