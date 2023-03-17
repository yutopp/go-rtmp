# go-rtmp

[![license](https://img.shields.io/github/license/livekit/go-rtmp.svg)](https://github.com/livekit/go-rtmp/blob/master/LICENSE_1_0.txt)

RTMP 1.0 server/client library written in Go, forked from [github.com/yutopp/go-rtmp](https://github.com/yutopp/go-rtmp) to change some of the error handling behavior.

## Installation

```
go get github.com/livekit/go-rtmp
```

See also [server_demo](https://github.com/livekit/go-rtmp/tree/master/example/server_demo) and [client_demo](https://github.com/livekit/go-rtmp/blob/master/example/client_demo/main.go).

## Documentation

- [GoDoc](https://pkg.go.dev/github.com/yutopp/go-rtmp)
- [REAL-TIME MESSAGING PROTOCOL (RTMP) SPECIFICATION](https://www.adobe.com/devnet/rtmp.html)


## NOTES

### How to limit bitrates or set timeouts

- Please use [yutopp/go-iowrap](https://github.com/yutopp/go-iowrap).

## License

[Boost Software License - Version 1.0](./LICENSE_1_0.txt)
