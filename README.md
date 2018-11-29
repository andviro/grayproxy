# grayproxy

A simple forwarding proxy for Graylog GELF messages. Can be used as
load-balancer in front of multiple Graylog server nodes. Also helps if the
Graylog server or logging application is behind a firewall and direct logging
to UDP is not viable.

## Usage

The basic usage for single UDP input and single GELF HTTP output is:

```
docker run --rm -p 12201:12201/udp -it andviro/grayproxy -out http://some.host/gelf
```

By default grayproxy configures the input at udp://0.0.0.0:12201 and no
outputs. Outputs are added using `-out` flag and may be specified multiple
times. Outputs may use HTTP, HTTPS, TCP and WebSocket protocol and are tried in
round-robin fashion. If message was not sent to any output, it will be silently
dropped unless disk buffer directory is configured. To listen on multiple TCP,
HTTP or UDP inputs, `-in` flag can be used.

## WebSocket output

```
-out ws://MyToken@127.0.0.1:20222
```

Emits all incoming messages to all connected WebSocket clients.
Message is firstly converted to flattened json structure.
Server side filtering of message is possible by specifying filters in query string.

```
curl --include \
     --no-buffer \
     --header "Connection: Upgrade" \
     --header "Upgrade: websocket" \
     --header "Host: 127.0.0.1:20222" \
     --header "Origin: http://127.0.0.1:20222" \
     --header "Sec-WebSocket-Key: SGVabG8sIHOvcmxDIQ==" \
     --header "Sec-WebSocket-Version: 13" \
     "http://127.0.0.1:20222/filter?token=MyToken&container_name=mycontainer"
```

## Command-line options

```
  -dataDir string
    	buffer directory (defaults to no buffering)
  -in value
    	input address in form schema://address:port (may be specified multiple times). Default: udp://:12201
  -out value
    	output address in form schema://address:port (may be specified multiple times)
  -sendTimeout int
    	maximum TCP or HTTP output timeout (ms) (default 1000)
```

## Credits

The idea of a pipelined GELF dechunker borrowed from [timtkachenko/gelf-go](https://github.com/timtkachenko/gelf-go).

## License

This code is released under 
[MIT](https://github.com/andviro/grayproxy/blob/master/LICENSE) license.
