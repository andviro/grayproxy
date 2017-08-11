FROM golang:1.8-alpine
RUN apk update && apk add git
RUN mkdir -p /go/src/github.com/andviro/grayproxy
WORKDIR /go/src/github.com/andviro/grayproxy
COPY . .
RUN go get
RUN go install
EXPOSE 12201/udp
ENTRYPOINT ["/go/bin/grayproxy"]
