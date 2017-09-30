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
times. Outputs may use HTTP, HTTPS and TCP protocol and are tried in
round-robin fashion. If message was not sent to any output, it is printed on
stdout. To listen on multiple TCP, HTTP or UDP inputs, `-in` flag can be used.

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
