# go-rtmp

[![ci](https://github.com/yutopp/go-rtmp/workflows/ci/badge.svg)](https://github.com/yutopp/go-rtmp/actions?query=workflow%3Aci)
[![codecov](https://codecov.io/gh/yutopp/go-rtmp/branch/master/graph/badge.svg?token=KXgQ1x8BQP)](https://codecov.io/gh/yutopp/go-rtmp)
[![GoDoc](https://godoc.org/github.com/yutopp/go-rtmp?status.svg)](http://godoc.org/github.com/yutopp/go-rtmp)
[![Go Report Card](https://goreportcard.com/badge/github.com/yutopp/go-rtmp)](https://goreportcard.com/report/github.com/yutopp/go-rtmp)
[![license](https://img.shields.io/github/license/yutopp/go-rtmp.svg)](https://github.com/yutopp/go-rtmp/blob/master/LICENSE_1_0.txt)

RTMP 1.0 server/client library written in Go.

*Work in progress*

## Installation

```
go get github.com/yutopp/go-rtmp
```

See also [server_demo](https://github.com/yutopp/go-rtmp/tree/master/example/server_demo) and [client_demo](https://github.com/yutopp/go-rtmp/blob/master/example/client_demo/main.go).

## Documentation

- [GoDoc](https://pkg.go.dev/github.com/yutopp/go-rtmp)
- [REAL-TIME MESSAGING PROTOCOL (RTMP) SPECIFICATION](https://www.adobe.com/devnet/rtmp.html)


## NOTES

### How to limit bitrates or set timeouts

- Please use [yutopp/go-iowrap](https://github.com/yutopp/go-iowrap).

## License

[Boost Software License - Version 1.0](./LICENSE_1_0.txt)
